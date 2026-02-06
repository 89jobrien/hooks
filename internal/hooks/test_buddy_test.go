package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTestBuddy_NudgesWhenNoTest(t *testing.T) {
	dir := t.TempDir()
	// Write a Go file with no corresponding test
	goFile := filepath.Join(dir, "handler.go")
	os.WriteFile(goFile, []byte("package main"), 0644)

	result, code := TestBuddy(writeInput(goFile, "package main"), dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Message == "" {
		t.Error("expected nudge message when test file missing")
	}
}

func TestTestBuddy_QuietWhenTestExists(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "handler.go")
	testFile := filepath.Join(dir, "handler_test.go")
	os.WriteFile(goFile, []byte("package main"), 0644)
	os.WriteFile(testFile, []byte("package main"), 0644)

	result, code := TestBuddy(writeInput(goFile, "package main"), dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Message != "" {
		t.Errorf("expected no message when test exists, got %q", result.Message)
	}
}

func TestTestBuddy_IgnoresTestFiles(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "handler_test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)

	result, code := TestBuddy(writeInput(testFile, "package main"), dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Message != "" {
		t.Errorf("expected no message for test file itself, got %q", result.Message)
	}
}

func TestTestBuddy_PythonFiles(t *testing.T) {
	dir := t.TempDir()
	pyFile := filepath.Join(dir, "handler.py")
	os.WriteFile(pyFile, []byte("def handle(): pass"), 0644)

	result, _ := TestBuddy(writeInput(pyFile, "def handle(): pass"), dir)
	if result.Message == "" {
		t.Error("expected nudge for .py file with no test_handler.py")
	}
}

func TestTestBuddy_PythonTestExists(t *testing.T) {
	dir := t.TempDir()
	pyFile := filepath.Join(dir, "handler.py")
	testFile := filepath.Join(dir, "test_handler.py")
	os.WriteFile(pyFile, []byte("def handle(): pass"), 0644)
	os.WriteFile(testFile, []byte("def test_handle(): pass"), 0644)

	result, _ := TestBuddy(writeInput(pyFile, "def handle(): pass"), dir)
	if result.Message != "" {
		t.Errorf("expected no message when Python test exists, got %q", result.Message)
	}
}

func TestTestBuddy_JSFiles(t *testing.T) {
	dir := t.TempDir()
	jsFile := filepath.Join(dir, "utils.ts")
	os.WriteFile(jsFile, []byte("export const x = 1"), 0644)

	result, _ := TestBuddy(writeInput(jsFile, "export const x = 1"), dir)
	if result.Message == "" {
		t.Error("expected nudge for .ts file with no test")
	}
}

func TestTestBuddy_JSTestExists(t *testing.T) {
	dir := t.TempDir()
	jsFile := filepath.Join(dir, "utils.ts")
	testFile := filepath.Join(dir, "utils.test.ts")
	os.WriteFile(jsFile, []byte("export const x = 1"), 0644)
	os.WriteFile(testFile, []byte("test('x', () => {})"), 0644)

	result, _ := TestBuddy(writeInput(jsFile, "export const x = 1"), dir)
	if result.Message != "" {
		t.Errorf("expected no message when JS test exists, got %q", result.Message)
	}
}

func TestTestBuddy_IgnoresNonCode(t *testing.T) {
	result, code := TestBuddy(writeInput("README.md", "# Hi"), "")
	if code != 0 || result.Message != "" {
		t.Errorf("should ignore non-code files")
	}
}

func TestTestBuddy_PassthroughNonWrite(t *testing.T) {
	result, code := TestBuddy(shellInput("ls"), "")
	if code != 0 || result.Decision != "allow" {
		t.Error("should passthrough non-Write tools")
	}
}
