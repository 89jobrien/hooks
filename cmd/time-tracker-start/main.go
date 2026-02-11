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
	if hooks.IsHookDisabled("time-tracker-start") {
		fmt.Println(`{}`)
		os.Exit(0)
	}
	data, _ := io.ReadAll(os.Stdin)
	var input hooks.HookInput
	json.Unmarshal(data, &input)

	home, _ := os.UserHomeDir()
	logDir := os.Getenv("HOOK_TIME_DIR")
	if logDir == "" {
		logDir = filepath.Join(home, ".cursor", "time")
	}

	result, code := hooks.TimeTracker(input, "start", logDir)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
