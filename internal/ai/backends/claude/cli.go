package claude

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"clawbench/internal/ai"
	"clawbench/internal/ai/backends"
	"clawbench/internal/model"
)

func init() {
	ai.RegisterBackend("claude", newClaudeBackend, true)
	backends.Register(&backends.BackendPlugin{
		ID: "claude",
		Spec: model.BackendSpec{
			ID: "claude", Backend: "claude", DefaultCmd: "claude", Name: "Claude", Icon: "🤖", Specialty: "代码编写与推理",
			DiscoverModelsFunc:   model.DiscoverClaudeModels,
			ThinkingEffortLevels: []string{"low", "medium", "high", "xhigh", "max"},
			AcpCommand:           "npx -y @agentclientprotocol/claude-agent-acp@latest",
			SortOrder:            1,
		},
		ACP: &backends.ACPPlugin{
			InputRemaps: ClaudeACPRemaps,
		},
	})
}

// newClaudeBackend returns a CLIBackend instance configured for Claude CLI.
func newClaudeBackend() ai.AIBackend {
	// cleanup holds the temp file cleanup function from the last BuildArgsFn call.
	// Safe because each agent has its own CLIBackend instance and sessions are sequential.
	var cleanup func()

	return &ai.CLIBackend{
		BackendName: "claude",
		Cmd:         "claude",
		BuildArgsFn: func(req ai.ChatRequest) []string {
			args, c := buildClaudeStreamArgs(req)
			cleanup = c
			return args
		},
		CleanupFn: func() {
			if cleanup != nil {
				cleanup()
				cleanup = nil
			}
		},
		NewParserFn:  func() ai.LineParser { return &ai.StreamParser{} },
		FilterLineFn: nil, // skip empty lines only (default)
		PreStartFn: func(cmd *exec.Cmd, req ai.ChatRequest) {
			// Claude CLI in --print mode with stdout piped (non-TTY) requires prompt
			// via stdin — positional prompt argument is not recognized.
			// Both new sessions and resume sessions use stdin for prompt.
			cmd.Stdin = strings.NewReader(req.Prompt)
		},
	}
}

// toClaudeAlias converts full Claude model IDs to CLI aliases.
// Claude CLI only accepts aliases (sonnet, opus, haiku), not full IDs (claude-sonnet-4-6).
var claudeModelAliases = map[string]string{
	"claude-sonnet-4-6": "sonnet",
	"claude-opus-4-6":   "opus",
	"claude-haiku-4-5":  "haiku",
	"claude-opus-4-8":   "opus",
	"claude-sonnet-4-5": "sonnet",
}

func toClaudeAlias(modelID string) string {
	if alias, ok := claudeModelAliases[modelID]; ok {
		return alias
	}
	// If already an alias or unknown, return as-is
	return modelID
}

// buildClaudeStreamArgs constructs CLI arguments for Claude streaming.
// When a system prompt is provided, it writes the prompt to a temp file and uses
// --system-prompt-file instead of --system-prompt to avoid Windows command-line
// length limits (error: "The filename or extension is too long").
func buildClaudeStreamArgs(req ai.ChatRequest) ([]string, func()) {
	args := ai.BuildBaseStreamArgs(req, func(r ai.ChatRequest) []string {
		extra := []string{"--verbose"}
		if r.Model != "" {
			extra = append(extra, "--model", toClaudeAlias(r.Model))
		}
		return extra
	})

	// No system prompt → no temp file needed.
	if req.SystemPrompt == "" {
		return args, nil
	}

	// Write system prompt to temp file and replace --system-prompt with --system-prompt-file.
	tmpFile, err := os.CreateTemp("", "clawbench-sysprompt-*.txt")
	if err != nil {
		slog.Error("claude: failed to create temp file for system prompt, falling back to inline",
			slog.String("error", err.Error()))
		return args, nil
	}
	if _, err := tmpFile.WriteString(req.SystemPrompt); err != nil {
		slog.Error("claude: failed to write system prompt temp file",
			slog.String("error", err.Error()))
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return args, nil
	}
	tmpFile.Close()

	// Replace --system-prompt <content> with --system-prompt-file <path> in args.
	var filtered []string
	skipNext := false
	for i, a := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if a == "--system-prompt" && i+1 < len(args) {
			filtered = append(filtered, "--system-prompt-file", tmpFile.Name())
			skipNext = true
			continue
		}
		filtered = append(filtered, a)
	}

	slog.Debug("claude: system prompt written to temp file",
		slog.String("path", tmpFile.Name()),
		slog.Int("size", len(req.SystemPrompt)))

	cleanupFn := func() {
		os.Remove(tmpFile.Name())
	}
	return filtered, cleanupFn
}
