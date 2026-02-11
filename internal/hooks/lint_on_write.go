package hooks

import (
	"fmt"
	"path/filepath"
)

var lintCommands = map[string]string{
	".go":   "gofmt -w %s",
	".py":   "ruff check --fix %s",
	".js":   "eslint --fix %s",
	".jsx":  "eslint --fix %s",
	".ts":   "eslint --fix %s",
	".tsx":  "eslint --fix %s",
	".rs":   "rustfmt %s",
	".sh":   "shfmt -w %s",
	".bash": "shfmt -w %s",
	".tf":   "terraform fmt %s",
}

// LintOnWrite is a postToolUse hook that suggests a lint command after file writes.
// Always exits 0 (informational only).
func LintOnWrite(input HookInput) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	path := input.Path()
	if path == "" {
		return Allow(), 0
	}

	ext := filepath.Ext(path)
	if tmpl, ok := lintCommands[ext]; ok {
		result := Allow()
		result.LintCommand = fmt.Sprintf(tmpl, path)
		return result, 0
	}

	return Allow(), 0
}
