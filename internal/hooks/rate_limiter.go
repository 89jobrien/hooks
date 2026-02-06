package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RateLimiter is a preToolUse hook that blocks excessive tool calls.
func RateLimiter(input HookInput, maxPerMinute int, stateDir string) (HookResult, int) {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return Allow(), 0
	}

	stateFile := filepath.Join(stateDir, "rate-limiter.state")
	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// Read existing timestamps
	var recent []time.Time
	if data, err := os.ReadFile(stateFile); err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			ts, err := time.Parse(time.RFC3339Nano, line)
			if err == nil && ts.After(cutoff) {
				recent = append(recent, ts)
			}
		}
	}

	// Check rate
	if len(recent) >= maxPerMinute {
		return Deny(fmt.Sprintf("Blocked: rate limit exceeded (%d calls in last minute, limit: %d). Possible runaway loop.", len(recent), maxPerMinute)), 2
	}

	// Record this call
	recent = append(recent, now)
	var sb strings.Builder
	for _, ts := range recent {
		sb.WriteString(ts.Format(time.RFC3339Nano))
		sb.WriteString("\n")
	}
	os.WriteFile(stateFile, []byte(sb.String()), 0644)

	return Allow(), 0
}
