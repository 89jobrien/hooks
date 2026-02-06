package hooks

import (
	"testing"
)

func TestLintOnWrite_SuggestsCorrectLinter(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{".go -> gofmt", "main.go", "gofmt"},
		{".py -> ruff", "app.py", "ruff"},
		{".js -> eslint", "index.js", "eslint"},
		{".ts -> eslint", "app.ts", "eslint"},
		{".tsx -> eslint", "App.tsx", "eslint"},
		{".jsx -> eslint", "App.jsx", "eslint"},
		{".rs -> rustfmt", "main.rs", "rustfmt"},
		{".sh -> shfmt", "deploy.sh", "shfmt"},
		{".tf -> terraform", "main.tf", "terraform"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := LintOnWrite(writeInput(tt.path, "content"))
			if code != 0 {
				t.Errorf("expected exit 0, got %d", code)
			}
			if result.Decision != "allow" {
				t.Errorf("expected allow, got %q", result.Decision)
			}
			if result.LintCommand == "" {
				t.Errorf("expected lint_command for %s, got empty", tt.path)
			}
			if !contains(result.LintCommand, tt.expected) {
				t.Errorf("expected lint_command containing %q, got %q", tt.expected, result.LintCommand)
			}
		})
	}
}

func TestLintOnWrite_NoHintForNonLintable(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"README.md", "README.md"},
		{"data.csv", "data.csv"},
		{"config.json", "config.json"},
		{"image.png", "image.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := LintOnWrite(writeInput(tt.path, "content"))
			if code != 0 {
				t.Errorf("expected exit 0, got %d", code)
			}
			if result.LintCommand != "" {
				t.Errorf("expected no lint_command for %s, got %q", tt.path, result.LintCommand)
			}
		})
	}
}

func TestLintOnWrite_PassthroughNonWrite(t *testing.T) {
	result, code := LintOnWrite(shellInput("ls"))
	if code != 0 || result.Decision != "allow" {
		t.Errorf("non-Write tool should passthrough")
	}
	if result.LintCommand != "" {
		t.Errorf("expected no lint_command for non-Write tool")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
