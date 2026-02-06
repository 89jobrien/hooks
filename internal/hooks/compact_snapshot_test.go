package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompactSnapshot_SavesContext(t *testing.T) {
	auditDir := t.TempDir()
	snapshotDir := t.TempDir()

	// Create a fake audit log
	lines := []string{
		"[2026-02-05 10:00:00] tool=Shell command=go test ./...",
		"[2026-02-05 10:01:00] tool=Write path=main.go",
		"[2026-02-05 10:02:00] tool=Shell command=go build ./...",
	}
	today := "2026-02-05"
	os.WriteFile(
		filepath.Join(auditDir, "audit-"+today+".log"),
		[]byte(strings.Join(lines, "\n")+"\n"),
		0644,
	)

	result, code := CompactSnapshot(HookInput{}, auditDir, snapshotDir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}

	entries, _ := os.ReadDir(snapshotDir)
	if len(entries) == 0 {
		t.Fatal("expected snapshot file to be created")
	}

	data, _ := os.ReadFile(filepath.Join(snapshotDir, entries[0].Name()))
	content := string(data)
	if !strings.Contains(content, "go test") {
		t.Errorf("snapshot missing audit content")
	}
}

func TestCompactSnapshot_HandlesEmpty(t *testing.T) {
	result, code := CompactSnapshot(HookInput{}, t.TempDir(), t.TempDir())
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}
