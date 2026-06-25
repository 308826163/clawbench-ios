//nolint:goconst // JSON response field names are domain strings, not config constants
package handler

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Skill represents a Claude Code skill.
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Enabled     bool   `json:"enabled"`
	Source      string `json:"source"` // "global", "plugin_cache", "plugin_market", "project"
	Plugin      string `json:"plugin,omitempty"`
}

// SkillListResponse is the response for GET /api/skills.
type SkillListResponse struct {
	Skills []Skill `json:"skills"`
	Total  int     `json:"total"`
}

// ServeSkills handles /api/skills requests.
func ServeSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		serveSkillsGet(w, r)
		return
	}
	if r.Method == http.MethodPost {
		serveSkillsPost(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		serveSkillsDelete(w, r)
		return
	}
	writeLocalizedErrorf(w, r, http.StatusMethodNotAllowed, "MethodNotAllowed")
}

// serveSkillsGet returns all skills from all sources.
func serveSkillsGet(w http.ResponseWriter, _ *http.Request) {
	skills := getAllSkills()

	// Sort: enabled first, then by source, then by name
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Enabled != skills[j].Enabled {
			return skills[i].Enabled // enabled first
		}
		if skills[i].Source != skills[j].Source {
			return skills[i].Source < skills[j].Source
		}
		return skills[i].Name < skills[j].Name
	})

	writeJSON(w, http.StatusOK, SkillListResponse{
		Skills: skills,
		Total:  len(skills),
	})
}

// serveSkillsPost creates a new skill in the global skills directory.
func serveSkillsPost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Name == "" {
		writeLocalizedErrorf(w, r, http.StatusBadRequest, "InvalidRequestBody")
		return
	}

	// Validate name (kebab-case)
	if strings.Contains(req.Name, " ") {
		writeLocalizedErrorf(w, r, http.StatusBadRequest, "InvalidSkillName")
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		writeLocalizedErrorf(w, r, http.StatusInternalServerError, "InternalError")
		return
	}

	skillDir := filepath.Join(homeDir, ".claude", "skills", req.Name)
	skillFile := filepath.Join(skillDir, "SKILL.md")

	// Check if skill already exists
	if _, err := os.Stat(skillFile); err == nil {
		writeLocalizedErrorf(w, r, http.StatusConflict, "SkillAlreadyExists")
		return
	}

	// Create directory
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		slog.Error("failed to create skill directory", "path", skillDir, "error", err)
		writeLocalizedErrorf(w, r, http.StatusInternalServerError, "InternalError")
		return
	}

	// Build SKILL.md content
	content := "---\n"
	content += "name: " + req.Name + "\n"
	if req.Description != "" {
		content += "description: " + req.Description + "\n"
	}
	content += "---\n\n"
	if req.Content != "" {
		content += req.Content + "\n"
	}

	// Write file
	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		slog.Error("failed to write skill file", "path", skillFile, "error", err)
		writeLocalizedErrorf(w, r, http.StatusInternalServerError, "InternalError")
		return
	}

	writeJSON(w, http.StatusOK, Skill{
		Name:        req.Name,
		Description: req.Description,
		Path:        skillFile,
		Enabled:     true,
		Source:      "global",
	})
}

// serveSkillsDelete deletes a skill.
func serveSkillsDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path   string `json:"path"`
		Source string `json:"source"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Path == "" {
		writeLocalizedErrorf(w, r, http.StatusBadRequest, "InvalidRequestBody")
		return
	}

	// Only allow deleting global skills
	if req.Source != "global" {
		writeLocalizedErrorf(w, r, http.StatusForbidden, "CannotDeletePluginSkill")
		return
	}

	// Get the skill directory (parent of SKILL.md)
	skillDir := filepath.Dir(req.Path)

	// Check if directory exists
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		writeLocalizedErrorf(w, r, http.StatusNotFound, "SkillNotFound")
		return
	}

	// Delete directory
	if err := os.RemoveAll(skillDir); err != nil {
		slog.Error("failed to delete skill directory", "path", skillDir, "error", err)
		writeLocalizedErrorf(w, r, http.StatusInternalServerError, "InternalError")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// getAllSkills scans all skill directories and returns a unified list.
func getAllSkills() []Skill {
	var skills []Skill

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return skills
	}

	// 1. Global skills (~/.claude/skills/)
	globalDir := filepath.Join(homeDir, ".claude", "skills")
	skills = append(skills, scanSkillDirectory(globalDir, "global", "")...)

	// 2. Plugin cache skills (~/.claude/plugins/cache/)
	cacheDir := filepath.Join(homeDir, ".claude", "plugins", "cache")
	skills = append(skills, scanPluginDirectory(cacheDir, "plugin_cache")...)

	// 3. Plugin market skills (~/.claude/plugins/marketplaces/)
	marketDir := filepath.Join(homeDir, ".claude", "plugins", "marketplaces")
	skills = append(skills, scanPluginDirectory(marketDir, "plugin_market")...)

	// 4. Project skills (current directory/.claude/skills/)
	projectDir := findProjectSkillsDir()
	if projectDir != "" {
		skills = append(skills, scanSkillDirectory(projectDir, "project", "")...)
	}

	// 5. Other project skills (scan common project directories)
	otherProjectDirs := findOtherProjectSkillsDirs()
	for _, dir := range otherProjectDirs {
		skills = append(skills, scanSkillDirectory(dir, "project", "")...)
	}

	return skills
}

// scanSkillDirectory scans a directory for SKILL.md files.
func scanSkillDirectory(dir, source, plugin string) []Skill {
	var skills []Skill

	entries, err := os.ReadDir(dir)
	if err != nil {
		return skills
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillFile := filepath.Join(dir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			continue
		}

		skill := parseSkillFile(skillFile, source, plugin)
		skill.Name = entry.Name()
		skills = append(skills, skill)
	}

	return skills
}

// scanPluginDirectory scans a plugin directory for skills.
func scanPluginDirectory(dir, source string) []Skill {
	var skills []Skill

	// Walk through plugin directories
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.Name() != "SKILL.md" {
			return nil
		}

		// Extract plugin name from path
		relPath, _ := filepath.Rel(dir, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		plugin := ""
		if len(parts) > 0 {
			plugin = parts[0]
		}

		skill := parseSkillFile(path, source, plugin)
		skill.Name = filepath.Base(filepath.Dir(path))
		skills = append(skills, skill)

		return nil
	})

	if err != nil {
		slog.Error("failed to walk plugin directory", "path", dir, "error", err)
	}

	return skills
}

// parseSkillFile reads a SKILL.md file and extracts metadata.
func parseSkillFile(path, source, plugin string) Skill {
	skill := Skill{
		Path:    path,
		Enabled: true,
		Source:  source,
		Plugin:  plugin,
	}

	// Check if disabled (directory ends with .disabled)
	dir := filepath.Dir(path)
	if strings.HasSuffix(dir, ".disabled") {
		skill.Enabled = false
		skill.Name = strings.TrimSuffix(filepath.Base(dir), ".disabled")
	}

	// Read file to extract description
	data, err := os.ReadFile(path)
	if err != nil {
		return skill
	}

	content := string(data)

	// Parse YAML frontmatter
	if strings.HasPrefix(content, "---") {
		lines := strings.Split(content, "\n")
		inFrontmatter := false
		for _, line := range lines {
			if line == "---" {
				if inFrontmatter {
					break
				}
				inFrontmatter = true
				continue
			}
			if inFrontmatter && strings.HasPrefix(line, "description:") {
				skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}

	return skill
}

// findProjectSkillsDir finds the .claude/skills directory in the current project.
func findProjectSkillsDir() string {
	// Try to find .claude/skills in current directory
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up to find .claude directory
	for {
		claudeDir := filepath.Join(dir, ".claude")
		if _, err := os.Stat(claudeDir); err == nil {
			skillsDir := filepath.Join(claudeDir, "skills")
			if _, err := os.Stat(skillsDir); err == nil {
				return skillsDir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// findOtherProjectSkillsDirs finds .claude/skills directories in other common project locations.
func findOtherProjectSkillsDirs() []string {
	var dirs []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return dirs
	}

	// Common project directories to scan
	commonDirs := []string{
		filepath.Join(homeDir, "Desktop"),
		filepath.Join(homeDir, "Documents"),
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "workspace"),
		filepath.Join(homeDir, "code"),
		filepath.Join(homeDir, "dev"),
	}

	// Also scan D:\ drive if on Windows
	if runtime.GOOS == "windows" {
		dDirs := []string{
			"D:\\AI",
			"D:\\Projects",
			"D:\\workspace",
			"D:\\code",
			"D:\\dev",
		}
		commonDirs = append(commonDirs, dDirs...)
	}

	// Scan each directory for .claude/skills
	for _, dir := range commonDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		// Walk through subdirectories
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			projectDir := filepath.Join(dir, entry.Name())
			claudeDir := filepath.Join(projectDir, ".claude")
			skillsDir := filepath.Join(claudeDir, "skills")

			if _, err := os.Stat(skillsDir); err == nil {
				dirs = append(dirs, skillsDir)
			}
		}
	}

	return dirs
}

// GetClaudeHomeDir returns the Claude home directory.
func GetClaudeHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude")
}
