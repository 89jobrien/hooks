package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// estimateTokens gives a rough token count for the hook input.
// Approximation: ~4 chars per token.
func estimateTokens(input HookInput) int {
	raw, _ := json.Marshal(input)
	return len(raw) / 4
}

// CostEstimator is a postToolUse hook that tracks estimated token usage.
func CostEstimator(input HookInput, logDir string) (HookResult, int) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return Allow(), 0
	}

	tokens := estimateTokens(input)
	logFile := filepath.Join(logDir, "cost.log")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return Allow(), 0
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] tool=%s tokens=%d\n", timestamp, input.ToolName, tokens)

	return Allow(), 0
}
