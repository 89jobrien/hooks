package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TimeTracker is a sessionStart/sessionEnd hook that logs session timestamps.
func TimeTracker(input HookInput, event string, logDir string) (HookResult, int) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return Allow(), 0
	}

	logFile := filepath.Join(logDir, "sessions.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return Allow(), 0
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	label := "START"
	if event == "end" {
		label = "END"
	}

	fmt.Fprintf(f, "[%s] %s\n", timestamp, label)
	return Allow(), 0
}
