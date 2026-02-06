package hooks

import (
	"os"
	"path/filepath"
	"strings"
)

// PromptEnricher is a beforeSubmitPrompt hook that injects project conventions.
// It looks for .cursor/conventions.md in the work directory.
func PromptEnricher(input HookInput, workDir string) (HookResult, int) {
	candidates := []string{
		filepath.Join(workDir, ".cursor", "conventions.md"),
		filepath.Join(workDir, ".claude", "conventions.md"),
		filepath.Join(workDir, "CONVENTIONS.md"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			content := strings.TrimSpace(string(data))
			if content != "" {
				return AllowMsg("[Project Conventions]\n" + content), 0
			}
		}
	}

	return Allow(), 0
}
