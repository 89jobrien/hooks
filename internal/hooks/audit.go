package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Audit is a postToolUse hook that logs all tool usage to an audit file.
// It always returns allow (never blocks).
func Audit(input HookInput, auditDir string) (HookResult, int) {
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return Allow(), 0
	}

	logFile := filepath.Join(auditDir, fmt.Sprintf("audit-%s.log", time.Now().Format("2006-01-02")))
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	summary := buildSummary(input)
	line := fmt.Sprintf("[%s] tool=%s %s\n", timestamp, input.ToolName, summary)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return Allow(), 0
	}
	defer f.Close()

	f.WriteString(line)
	return Allow(), 0
}

func buildSummary(input HookInput) string {
	switch input.ToolName {
	case "Shell":
		cmd := input.Command()
		if len(cmd) > 200 {
			cmd = cmd[:200]
		}
		return "command=" + cmd
	case "Write", "Read":
		return "path=" + input.Path()
	case "Grep":
		return "pattern=" + input.Pattern()
	default:
		return ""
	}
}
