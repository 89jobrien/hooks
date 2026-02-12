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
	if hooks.IsHookDisabled("dry-run-mode") {
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

	enabled := os.Getenv("HOOKS_DRY_RUN") == "1"

	home, _ := os.UserHomeDir()
	logDir := os.Getenv("HOOK_DRY_RUN_DIR")
	if logDir == "" {
		logDir = filepath.Join(home, ".config", "hooks", "dry-run")
	}

	result, code := hooks.DryRunMode(input, enabled, logDir)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
