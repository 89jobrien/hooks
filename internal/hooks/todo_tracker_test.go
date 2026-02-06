package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTodoTracker_DetectsTodos(t *testing.T) {
	dir := t.TempDir()

	contents := "package main\n// TODO: fix this\nfunc main() {\n\t// FIXME: broken\n\t// HACK: workaround\n}\n"
	result, code := TodoTracker(writeInput("main.go", contents), dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}

	logFile := filepath.Join(dir, "TODO.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected TODO.log to be created: %v", err)
	}

	log := string(data)
	if !strings.Contains(log, "TODO") {
		t.Errorf("log missing TODO, got: %s", log)
	}
	if !strings.Contains(log, "FIXME") {
		t.Errorf("log missing FIXME, got: %s", log)
	}
	if !strings.Contains(log, "HACK") {
		t.Errorf("log missing HACK, got: %s", log)
	}
}

func TestTodoTracker_NoTodosNoLog(t *testing.T) {
	dir := t.TempDir()

	contents := "package main\nfunc main() {}\n"
	result, code := TodoTracker(writeInput("main.go", contents), dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Message != "" {
		t.Errorf("expected no message for clean file, got %q", result.Message)
	}
}

func TestTodoTracker_IncludesLineNumbers(t *testing.T) {
	dir := t.TempDir()

	contents := "line1\nline2\n// TODO: do something\nline4\n"
	TodoTracker(writeInput("app.go", contents), dir)

	data, _ := os.ReadFile(filepath.Join(dir, "TODO.log"))
	if !strings.Contains(string(data), "app.go:3") {
		t.Errorf("expected line number in log, got: %s", string(data))
	}
}

func TestTodoTracker_PassthroughNonWrite(t *testing.T) {
	result, code := TodoTracker(shellInput("ls"), "")
	if code != 0 || result.Decision != "allow" {
		t.Error("should passthrough non-Write tools")
	}
}
