package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter_AllowsNormal(t *testing.T) {
	dir := t.TempDir()
	result, code := RateLimiter(shellInput("ls"), 30, dir)
	if code != 0 {
		t.Errorf("expected allow, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}

func TestRateLimiter_BlocksExcessive(t *testing.T) {
	dir := t.TempDir()

	// Simulate 31 calls in the last minute
	stateFile := filepath.Join(dir, "rate-limiter.state")
	var lines []string
	now := time.Now()
	for i := 0; i < 31; i++ {
		ts := now.Add(-time.Duration(i) * time.Second).Format(time.RFC3339Nano)
		lines = append(lines, ts)
	}
	os.WriteFile(stateFile, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	result, code := RateLimiter(shellInput("ls"), 30, dir)
	if code != 2 {
		t.Errorf("expected block (exit 2), got %d", code)
	}
	if result.Decision != "deny" {
		t.Errorf("expected deny, got %q", result.Decision)
	}
}

func TestRateLimiter_PrunesOldEntries(t *testing.T) {
	dir := t.TempDir()

	// Simulate old entries (>1 minute ago)
	stateFile := filepath.Join(dir, "rate-limiter.state")
	var lines []string
	old := time.Now().Add(-2 * time.Minute)
	for i := 0; i < 50; i++ {
		ts := old.Add(-time.Duration(i) * time.Second).Format(time.RFC3339Nano)
		lines = append(lines, ts)
	}
	os.WriteFile(stateFile, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	result, code := RateLimiter(shellInput("ls"), 30, dir)
	if code != 0 {
		t.Errorf("old entries should be pruned; expected allow, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}

func TestRateLimiter_PassthroughNonShell(t *testing.T) {
	// Rate limiter applies to ALL tool calls
	dir := t.TempDir()
	result, code := RateLimiter(writeInput("main.go", "x"), 30, dir)
	if code != 0 || result.Decision != "allow" {
		t.Error("should allow normal calls")
	}
}
