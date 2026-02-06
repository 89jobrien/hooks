package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CompactSnapshot is a preCompact hook that saves the audit log before compaction.
func CompactSnapshot(input HookInput, auditDir, snapshotDir string) (HookResult, int) {
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return Allow(), 0
	}

	today := time.Now().Format("2006-01-02")
	auditFile := filepath.Join(auditDir, "audit-"+today+".log")

	data, err := os.ReadFile(auditFile)
	if err != nil || len(data) == 0 {
		return AllowMsg("No audit data to snapshot"), 0
	}

	snapshotFile := filepath.Join(snapshotDir,
		fmt.Sprintf("pre-compact-%s.log", time.Now().Format("2006-01-02-150405")))

	header := fmt.Sprintf("# Pre-compaction snapshot — %s\n# Audit log at time of compaction:\n\n",
		time.Now().Format("2006-01-02 15:04:05"))

	os.WriteFile(snapshotFile, []byte(header+string(data)), 0644)

	return AllowMsg("Context snapshot saved before compaction"), 0
}
