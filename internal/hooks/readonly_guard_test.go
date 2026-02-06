package hooks

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReadonlyGuard_NonWriteTool(t *testing.T) {
	input := HookInput{
		ToolName: "Shell",
		ToolInput: []byte(`{"command": "ls"}`),
	}
	result, code := ReadonlyGuard(input)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-Write tool, got decision=%s code=%d", result.Decision, code)
	}
}

func TestReadonlyGuard_NoFilePath(t *testing.T) {
	input := writeInput("", "content")
	result, code := ReadonlyGuard(input)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no file path, got decision=%s code=%d", result.Decision, code)
	}
}

func TestReadonlyGuard_LockFiles(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantDeny bool
	}{
		{"package-lock.json", "package-lock.json", true},
		{"yarn.lock", "yarn.lock", true},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", true},
		{"poetry.lock", "poetry.lock", true},
		{"Cargo.lock", "Cargo.lock", true},
		{"uv.lock", "uv.lock", true},
		{"Gemfile.lock", "Gemfile.lock", true},
		{"composer.lock", "composer.lock", true},
		{"regular file", "package.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := writeInput(tt.filePath, "content")
			result, code := ReadonlyGuard(input)
			if tt.wantDeny {
				if code != 2 || result.Decision != "deny" {
					t.Errorf("expected deny for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
				if !strings.Contains(result.Reason, "Readonly file protection") {
					t.Errorf("expected reason to mention protection, got %q", result.Reason)
				}
			} else {
				if code != 0 || result.Decision != "allow" {
					t.Errorf("expected allow for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
			}
		})
	}
}

func TestReadonlyGuard_GeneratedFiles(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantDeny bool
	}{
		{"minified JS", "app.min.js", true},
		{"minified CSS", "styles.min.css", true},
		{"source map", "app.js.map", true},
		{"TypeScript declarations", "types.d.ts", true},
		{"regular JS", "app.js", false},
		{"regular CSS", "styles.css", false},
		{"regular TS", "app.ts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := writeInput(tt.filePath, "content")
			result, code := ReadonlyGuard(input)
			if tt.wantDeny {
				if code != 2 || result.Decision != "deny" {
					t.Errorf("expected deny for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
			} else {
				if code != 0 || result.Decision != "allow" {
					t.Errorf("expected allow for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
			}
		})
	}
}

func TestReadonlyGuard_VendorDirectories(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantDeny bool
	}{
		{"node_modules", "node_modules/package/index.js", true},
		{"vendor", "vendor/composer/autoload.php", true},
		{"pycache", "__pycache__/module.pyc", true},
		{"git", ".git/config", true},
		{"dist", "dist/bundle.js", true},
		{"build", "build/output.js", true},
		{"next", ".next/server.js", true},
		{"nuxt", ".nuxt/router.js", true},
		{"regular file", "src/app.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := writeInput(tt.filePath, "content")
			result, code := ReadonlyGuard(input)
			if tt.wantDeny {
				if code != 2 || result.Decision != "deny" {
					t.Errorf("expected deny for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
			} else {
				if code != 0 || result.Decision != "allow" {
					t.Errorf("expected allow for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
			}
		})
	}
}

func TestReadonlyGuard_IDEFiles(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantDeny bool
	}{
		{".idea", ".idea/workspace.xml", true},
		{"vscode settings", ".vscode/settings.json", true},
		{"vscode launch", ".vscode/launch.json", false}, // Override allowed
		{"vscode tasks", ".vscode/tasks.json", false},   // Override allowed
		{"regular file", "src/app.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := writeInput(tt.filePath, "content")
			result, code := ReadonlyGuard(input)
			if tt.wantDeny {
				if code != 2 || result.Decision != "deny" {
					t.Errorf("expected deny for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
			} else {
				if code != 0 || result.Decision != "allow" {
					t.Errorf("expected allow for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
			}
		})
	}
}

func TestReadonlyGuard_EditTool(t *testing.T) {
	ti, _ := json.Marshal(map[string]string{
		"file_path": "package-lock.json",
	})
	input := HookInput{
		ToolName:  "Edit",
		ToolInput: ti,
	}

	result, code := ReadonlyGuard(input)
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny for Edit on lock file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestReadonlyGuard_MultiEditTool(t *testing.T) {
	ti, _ := json.Marshal(map[string]string{
		"file_path": "yarn.lock",
	})
	input := HookInput{
		ToolName:  "MultiEdit",
		ToolInput: ti,
	}

	result, code := ReadonlyGuard(input)
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny for MultiEdit on lock file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestReadonlyGuard_PathWithBackslashes(t *testing.T) {
	// Test Windows-style paths
	input := writeInput("node_modules\\package\\index.js", "content")
	result, code := ReadonlyGuard(input)
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny for Windows path, got decision=%s code=%d", result.Decision, code)
	}
}

func TestReadonlyGuard_ReasonContainsHints(t *testing.T) {
	input := writeInput("package-lock.json", "content")
	result, code := ReadonlyGuard(input)
	if code != 2 {
		t.Fatal("expected deny")
	}

	if !strings.Contains(result.Reason, "Lock files") {
		t.Error("expected reason to contain hint about lock files")
	}
	if !strings.Contains(result.Reason, "package manager") {
		t.Error("expected reason to mention package manager")
	}
}
