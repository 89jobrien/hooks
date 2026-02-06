package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAudit_AlwaysAllows(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name  string
		input HookInput
	}{
		{"Shell tool", shellInput("ls -la")},
		{"Write tool", writeInput("main.go", "package main")},
		{"Read tool", func() HookInput {
			return HookInput{ToolName: "Read", ToolInput: []byte(`{"path":"README.md"}`)}
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := Audit(tt.input, dir)
			if code != 0 {
				t.Errorf("expected exit 0, got %d", code)
			}
			if result.Decision != "allow" {
				t.Errorf("expected allow, got %q", result.Decision)
			}
		})
	}
}

func TestAudit_CreatesLogFile(t *testing.T) {
	dir := t.TempDir()
	Audit(shellInput("git status"), dir)

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(dir, "audit-"+today+".log")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("log file is empty")
	}
}

func TestAudit_LogsToolName(t *testing.T) {
	dir := t.TempDir()
	Audit(shellInput("git status"), dir)

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(dir, "audit-"+today+".log")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	if !strings.Contains(string(data), "tool=Shell") {
		t.Errorf("log missing tool name, got: %s", string(data))
	}
}

func TestAudit_LogsCommand(t *testing.T) {
	dir := t.TempDir()
	Audit(shellInput("git status"), dir)

	today := time.Now().Format("2006-01-02")
	data, _ := os.ReadFile(filepath.Join(dir, "audit-"+today+".log"))

	if !strings.Contains(string(data), "git status") {
		t.Errorf("log missing command, got: %s", string(data))
	}
}

func TestAudit_LogsWritePath(t *testing.T) {
	dir := t.TempDir()
	Audit(writeInput("main.go", "package main"), dir)

	today := time.Now().Format("2006-01-02")
	data, _ := os.ReadFile(filepath.Join(dir, "audit-"+today+".log"))

	if !strings.Contains(string(data), "main.go") {
		t.Errorf("log missing path, got: %s", string(data))
	}
}

func TestAudit_LogsTimestamp(t *testing.T) {
	dir := t.TempDir()
	Audit(shellInput("echo test"), dir)

	today := time.Now().Format("2006-01-02")
	data, _ := os.ReadFile(filepath.Join(dir, "audit-"+today+".log"))

	if !strings.Contains(string(data), today) {
		t.Errorf("log missing timestamp, got: %s", string(data))
	}
}

func TestAudit_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	Audit(shellInput("first command"), dir)
	Audit(shellInput("second command"), dir)

	today := time.Now().Format("2006-01-02")
	data, _ := os.ReadFile(filepath.Join(dir, "audit-"+today+".log"))

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log lines, got %d: %s", len(lines), string(data))
	}
}
