package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDryRunMode_BlocksWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	result, code := DryRunMode(shellInput("npm install express"), true, dir)
	if code != 2 {
		t.Errorf("expected block (exit 2) in dry run, got %d", code)
	}
	if result.Decision != "deny" {
		t.Errorf("expected deny, got %q", result.Decision)
	}
	if !strings.Contains(result.Reason, "dry run") && !strings.Contains(result.Reason, "DRY RUN") {
		t.Errorf("reason should mention dry run, got: %s", result.Reason)
	}
}

func TestDryRunMode_LogsCommand(t *testing.T) {
	dir := t.TempDir()
	DryRunMode(shellInput("go test ./..."), true, dir)

	data, err := os.ReadFile(filepath.Join(dir, "dry-run.log"))
	if err != nil {
		t.Fatalf("expected dry-run.log: %v", err)
	}
	if !strings.Contains(string(data), "go test") {
		t.Errorf("log missing command, got: %s", string(data))
	}
}

func TestDryRunMode_AllowsWhenDisabled(t *testing.T) {
	result, code := DryRunMode(shellInput("npm install"), false, "")
	if code != 0 {
		t.Errorf("expected allow when dry run disabled, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}

func TestDryRunMode_PassthroughNonShell(t *testing.T) {
	result, code := DryRunMode(writeInput("main.go", "x"), true, "")
	if code != 0 || result.Decision != "allow" {
		t.Error("should passthrough non-Shell even in dry run")
	}
}

func TestDryRunMode_BlocksAllShellInDryRun(t *testing.T) {
	dir := t.TempDir()
	cmds := []string{"ls -la", "git status", "make build", "docker ps"}
	for _, cmd := range cmds {
		result, code := DryRunMode(shellInput(cmd), true, dir)
		if code != 2 {
			t.Errorf("expected block for %q in dry run, got %d", cmd, code)
		}
		if result.Decision != "deny" {
			t.Errorf("expected deny for %q, got %q", cmd, result.Decision)
		}
	}
}
