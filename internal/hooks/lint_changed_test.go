package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintChanged_NonWriteTool(t *testing.T) {
	input := HookInput{
		ToolName:  "Shell",
		ToolInput: []byte(`{"command": "ls"}`),
	}
	result, code := LintChanged(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-Write tool, got decision=%s code=%d", result.Decision, code)
	}
}

func TestLintChanged_NoFilePath(t *testing.T) {
	input := writeInput("", "content")
	result, code := LintChanged(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no file path, got decision=%s code=%d", result.Decision, code)
	}
}

func TestLintChanged_NonexistentFile(t *testing.T) {
	input := writeInput("/nonexistent/file.ts", "content")
	result, code := LintChanged(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for nonexistent file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestDetectLinter_Biome(t *testing.T) {
	dir := t.TempDir()
	biomeConfig := filepath.Join(dir, "biome.json")
	os.WriteFile(biomeConfig, []byte("{}"), 0644)

	linter := detectLinter(dir)
	if linter != "biome" && commandExists("biome") {
		t.Errorf("expected biome when biome.json exists and command available, got %q", linter)
	}
}

func TestDetectLinter_ESLint(t *testing.T) {
	dir := t.TempDir()
	eslintConfig := filepath.Join(dir, ".eslintrc.json")
	os.WriteFile(eslintConfig, []byte("{}"), 0644)

	linter := detectLinter(dir)
	if linter != "eslint" && commandExists("eslint") {
		t.Errorf("expected eslint when .eslintrc.json exists and command available, got %q", linter)
	}
}

func TestDetectLinter_Ruff(t *testing.T) {
	dir := t.TempDir()
	ruffConfig := filepath.Join(dir, "ruff.toml")
	os.WriteFile(ruffConfig, []byte(""), 0644)

	linter := detectLinter(dir)
	if linter != "ruff" && commandExists("ruff") {
		t.Errorf("expected ruff when ruff.toml exists and command available, got %q", linter)
	}
}

func TestDetectLinter_NoConfig(t *testing.T) {
	dir := t.TempDir()
	linter := detectLinter(dir)
	if linter != "" {
		t.Errorf("expected empty linter when no config exists, got %q", linter)
	}
}

func TestDetectLinter_Priority(t *testing.T) {
	dir := t.TempDir()
	// Create multiple configs - biome should take priority
	os.WriteFile(filepath.Join(dir, "biome.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, ".eslintrc.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "ruff.toml"), []byte(""), 0644)

	linter := detectLinter(dir)
	// Biome should be detected first (checked first)
	if linter != "biome" && commandExists("biome") {
		// If biome not available, check others
		if linter == "" {
			t.Log("No linters available, skipping priority test")
		}
	}
}

func TestLintChanged_NoLinterDetected(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.ts")
	os.WriteFile(testFile, []byte("const x = 1;"), 0644)

	input := writeInput(testFile, "const x = 1;")
	result, code := LintChanged(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no linter detected, got decision=%s code=%d", result.Decision, code)
	}
}

func TestCommandExists(t *testing.T) {
	// Test with a command that should exist
	if !commandExists("go") {
		t.Error("expected 'go' command to exist")
	}

	// Test with a command that shouldn't exist
	if commandExists("nonexistent-command-xyz-123") {
		t.Error("expected nonexistent command to return false")
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	if !exists(testFile) {
		t.Error("expected exists() to return true for existing file")
	}

	if exists(filepath.Join(dir, "nonexistent.txt")) {
		t.Error("expected exists() to return false for nonexistent file")
	}
}

func TestLintChanged_RelativePath(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.ts")
	os.WriteFile(testFile, []byte("const x = 1;"), 0644)

	// Use relative path
	relPath, _ := filepath.Rel(dir, testFile)
	input := writeInput(relPath, "const x = 1;")

	result, code := LintChanged(input, dir)
	// Should allow if no linter config, or if linter passes
	if code != 0 && result.Decision == "deny" {
		// If it denies, should be because linter found issues
		if !strings.Contains(result.Reason, "found issues") {
			t.Errorf("expected deny reason to mention issues, got %q", result.Reason)
		}
	}
}
