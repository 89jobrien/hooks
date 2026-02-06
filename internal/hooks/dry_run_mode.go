package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DryRunMode is a preToolUse hook that blocks Shell commands when dry run is enabled.
// It logs what would have been executed.
func DryRunMode(input HookInput, enabled bool, logDir string) (HookResult, int) {
	if !enabled {
		return Allow(), 0
	}

	if input.ToolName != "Shell" {
		return Allow(), 0
	}

	cmd := input.Command()

	// Log the blocked command
	if logDir != "" {
		os.MkdirAll(logDir, 0755)
		logFile := filepath.Join(logDir, "dry-run.log")
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			fmt.Fprintf(f, "[%s] DRY RUN blocked: %s\n", timestamp, cmd)
		}
	}

	return Deny(fmt.Sprintf("DRY RUN: would execute: %s", cmd)), 2
}
