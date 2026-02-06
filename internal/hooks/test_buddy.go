package hooks

import (
	"os"
	"path/filepath"
	"strings"
)

// testableExts maps source extensions to their test file patterns.
type testPattern struct {
	// Returns candidate test file paths given the source file path (without dir)
	candidates func(base string) []string
}

var testableExts = map[string]testPattern{
	".go": {func(base string) []string {
		return []string{strings.TrimSuffix(base, ".go") + "_test.go"}
	}},
	".py": {func(base string) []string {
		name := strings.TrimSuffix(base, ".py")
		return []string{"test_" + name + ".py", name + "_test.py"}
	}},
	".js": {func(base string) []string {
		name := strings.TrimSuffix(base, ".js")
		return []string{name + ".test.js", name + ".spec.js"}
	}},
	".ts": {func(base string) []string {
		name := strings.TrimSuffix(base, ".ts")
		return []string{name + ".test.ts", name + ".spec.ts"}
	}},
	".tsx": {func(base string) []string {
		name := strings.TrimSuffix(base, ".tsx")
		return []string{name + ".test.tsx", name + ".spec.tsx"}
	}},
	".jsx": {func(base string) []string {
		name := strings.TrimSuffix(base, ".jsx")
		return []string{name + ".test.jsx", name + ".spec.jsx"}
	}},
	".rs": {func(base string) []string {
		// Rust tests are typically in the same file, but check for a test module file
		name := strings.TrimSuffix(base, ".rs")
		return []string{name + "_test.rs"}
	}},
}

// isTestFile returns true if the file looks like a test file.
func isTestFile(path string) bool {
	base := filepath.Base(path)
	lower := strings.ToLower(base)
	return strings.Contains(lower, "_test.") ||
		strings.Contains(lower, ".test.") ||
		strings.Contains(lower, ".spec.") ||
		strings.HasPrefix(lower, "test_")
}

// TestBuddy is a postToolUse hook that nudges creation of test files.
func TestBuddy(input HookInput, workDir string) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	path := input.Path()
	if path == "" {
		return Allow(), 0
	}

	ext := filepath.Ext(path)
	pattern, ok := testableExts[ext]
	if !ok {
		return Allow(), 0
	}

	// Don't nudge for test files themselves
	if isTestFile(path) {
		return Allow(), 0
	}

	base := filepath.Base(path)
	dir := filepath.Dir(path)
	if workDir != "" && !filepath.IsAbs(path) {
		dir = workDir
	}

	candidates := pattern.candidates(base)
	for _, c := range candidates {
		testPath := filepath.Join(dir, c)
		if _, err := os.Stat(testPath); err == nil {
			return Allow(), 0
		}
	}

	return HookResult{
		Decision: "allow",
		Message:  "No test file found for " + base + ". Consider creating " + candidates[0],
	}, 0
}
