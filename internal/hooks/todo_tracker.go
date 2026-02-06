package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var todoRe = regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX|BUG)\b[:\s].*`)

// TodoTracker is a postToolUse hook that detects TODO/FIXME/HACK comments.
func TodoTracker(input HookInput, logDir string) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	contents := input.Contents()
	path := input.Path()
	if contents == "" || path == "" {
		return Allow(), 0
	}

	lines := strings.Split(contents, "\n")
	var found []string

	for i, line := range lines {
		if match := todoRe.FindString(line); match != "" {
			found = append(found, fmt.Sprintf("%s:%d: %s", filepath.Base(path), i+1, strings.TrimSpace(match)))
		}
	}

	if len(found) == 0 {
		return Allow(), 0
	}

	// Write to log
	if logDir != "" {
		logFile := filepath.Join(logDir, "TODO.log")
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			for _, entry := range found {
				fmt.Fprintf(f, "[%s] %s\n", timestamp, entry)
			}
		}
	}

	return AllowMsg(fmt.Sprintf("Found %d TODO/FIXME/HACK comment(s)", len(found))), 0
}
