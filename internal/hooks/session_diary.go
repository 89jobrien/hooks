package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var auditLineRe = regexp.MustCompile(`\[([^\]]+)\]\s+tool=(\S+)\s*(.*)`)

// SessionDiary is a stop hook that summarizes the session from the audit log.
func SessionDiary(input HookInput, auditDir, diaryDir string) (HookResult, int) {
	if err := os.MkdirAll(diaryDir, 0755); err != nil {
		return Allow(), 0
	}

	today := time.Now().Format("2006-01-02")
	auditFile := filepath.Join(auditDir, "audit-"+today+".log")

	data, err := os.ReadFile(auditFile)
	if err != nil {
		return AllowMsg("No audit log found for today"), 0
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		return AllowMsg("Empty audit log"), 0
	}

	// Aggregate stats
	toolCounts := make(map[string]int)
	var filesWritten []string
	var commands []string

	for _, line := range lines {
		m := auditLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		tool := m[2]
		detail := m[3]
		toolCounts[tool]++

		switch tool {
		case "Write":
			if strings.HasPrefix(detail, "path=") {
				filesWritten = append(filesWritten, strings.TrimPrefix(detail, "path="))
			}
		case "Shell":
			if strings.HasPrefix(detail, "command=") {
				cmd := strings.TrimPrefix(detail, "command=")
				if len(cmd) > 80 {
					cmd = cmd[:80] + "..."
				}
				commands = append(commands, cmd)
			}
		}
	}

	// Build diary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Session Diary — %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	sb.WriteString("## Tool Usage\n")
	for tool, count := range toolCounts {
		sb.WriteString(fmt.Sprintf("- %s: %d calls\n", tool, count))
	}

	if len(filesWritten) > 0 {
		sb.WriteString("\n## Files Written\n")
		seen := make(map[string]bool)
		for _, f := range filesWritten {
			if !seen[f] {
				sb.WriteString(fmt.Sprintf("- %s\n", f))
				seen[f] = true
			}
		}
	}

	if len(commands) > 0 {
		sb.WriteString("\n## Commands Run\n")
		for _, c := range commands {
			sb.WriteString(fmt.Sprintf("- `%s`\n", c))
		}
	}

	// Write diary
	diaryFile := filepath.Join(diaryDir, fmt.Sprintf("session-%s.md", time.Now().Format("2006-01-02-150405")))
	os.WriteFile(diaryFile, []byte(sb.String()), 0644)

	return AllowMsg(fmt.Sprintf("Session diary written: %d tool calls, %d files", len(lines), len(filesWritten))), 0
}
