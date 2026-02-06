package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionDiary_CreatesEntry(t *testing.T) {
	auditDir := t.TempDir()
	diaryDir := t.TempDir()

	// Create a fake audit log
	today := "2026-02-05"
	auditLog := filepath.Join(auditDir, "audit-"+today+".log")
	lines := []string{
		"[2026-02-05 10:00:00] tool=Shell command=go test ./...",
		"[2026-02-05 10:00:05] tool=Write path=main.go",
		"[2026-02-05 10:00:10] tool=Shell command=git status",
		"[2026-02-05 10:00:15] tool=Read path=config.yaml",
		"[2026-02-05 10:00:20] tool=Write path=handler.go",
	}
	os.WriteFile(auditLog, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	result, code := SessionDiary(HookInput{}, auditDir, diaryDir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}

	// Check diary was created
	entries, _ := os.ReadDir(diaryDir)
	if len(entries) == 0 {
		t.Fatal("expected diary entry to be created")
	}

	data, _ := os.ReadFile(filepath.Join(diaryDir, entries[0].Name()))
	diary := string(data)

	if !strings.Contains(diary, "Shell") {
		t.Errorf("diary missing Shell tool info")
	}
	if !strings.Contains(diary, "Write") {
		t.Errorf("diary missing Write tool info")
	}
}

func TestSessionDiary_HandlesNoAuditLog(t *testing.T) {
	auditDir := t.TempDir()
	diaryDir := t.TempDir()

	result, code := SessionDiary(HookInput{}, auditDir, diaryDir)
	if code != 0 {
		t.Errorf("expected exit 0 even with no audit log, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}
