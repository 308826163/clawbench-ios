package ai

// buildClaudeStreamArgs constructs the CLI arguments for Claude streaming
func buildClaudeStreamArgs(req ChatRequest) []string {
	return BuildBaseStreamArgs(req, func(r ChatRequest) []string {
		return []string{"--verbose"}
	})
}
