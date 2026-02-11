package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTimeTracker_LogsStart(t *testing.T) {
	dir := t.TempDir()
	result, code := TimeTracker(HookInput{}, "start", dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for session hook, got %q", result.Decision)
	}

	data, err := os.ReadFile(filepath.Join(dir, "sessions.log"))
	if err != nil {
		t.Fatalf("expected sessions.log: %v", err)
	}
	if !strings.Contains(string(data), "START") {
		t.Errorf("log missing START, got: %s", string(data))
	}
}

func TestTimeTracker_LogsEnd(t *testing.T) {
	dir := t.TempDir()
	TimeTracker(HookInput{}, "start", dir)
	TimeTracker(HookInput{}, "end", dir)

	data, _ := os.ReadFile(filepath.Join(dir, "sessions.log"))
	if !strings.Contains(string(data), "END") {
		t.Errorf("log missing END, got: %s", string(data))
	}
}

func TestTimeTracker_AlwaysExitsZero(t *testing.T) {
	_, code := TimeTracker(HookInput{}, "start", t.TempDir())
	if code != 0 {
		t.Errorf("should always exit 0, got %d", code)
	}
}
