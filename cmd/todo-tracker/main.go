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
	if hooks.IsHookDisabled("todo-tracker") {
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

	logDir := os.Getenv("HOOK_TODO_DIR")
	if logDir == "" {
		home, _ := os.UserHomeDir()
		logDir = filepath.Join(home, ".config", "hooks", "todos")
	}

	result, code := hooks.TodoTracker(input, logDir)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
