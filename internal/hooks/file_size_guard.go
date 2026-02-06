package hooks

import (
	"fmt"
	"strings"
)

// FileSizeGuard is a preToolUse hook that blocks writes producing files over maxLines.
func FileSizeGuard(input HookInput, maxLines int) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	contents := input.Contents()
	if contents == "" {
		return Allow(), 0
	}

	lineCount := strings.Count(contents, "\n") + 1
	if lineCount > maxLines {
		return Deny(fmt.Sprintf("Blocked: file would be %d lines (limit: %d). Break it into smaller files.", lineCount, maxLines)), 2
	}

	return Allow(), 0
}
