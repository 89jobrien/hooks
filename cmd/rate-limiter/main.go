package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if hooks.IsHookDisabled("rate-limiter") {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}

	var input hooks.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}

	maxPerMinute := 30
	if v := os.Getenv("HOOK_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxPerMinute = n
		}
	}

	home, _ := os.UserHomeDir()
	stateDir := os.Getenv("HOOK_RATE_DIR")
	if stateDir == "" {
		stateDir = filepath.Join(home, ".cursor", "rate")
	}

	result, code := hooks.RateLimiter(input, maxPerMinute, stateDir)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
