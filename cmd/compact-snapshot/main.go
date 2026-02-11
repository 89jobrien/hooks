package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if hooks.IsHookDisabled("compact-snapshot") {
		fmt.Println(`{}`)
		os.Exit(0)
	}
	data, _ := io.ReadAll(os.Stdin)
	var input hooks.HookInput
	json.Unmarshal(data, &input)

	home, _ := os.UserHomeDir()
	auditDir := os.Getenv("HOOK_AUDIT_DIR")
	if auditDir == "" {
		auditDir = filepath.Join(home, ".cursor", "audit")
	}
	snapshotDir := os.Getenv("HOOK_SNAPSHOT_DIR")
	if snapshotDir == "" {
		snapshotDir = filepath.Join(home, ".cursor", "snapshots")
	}

	result, code := hooks.CompactSnapshot(input, auditDir, snapshotDir)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
