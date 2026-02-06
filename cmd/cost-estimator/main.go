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
	if hooks.IsHookDisabled("cost-estimator") {
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

	home, _ := os.UserHomeDir()
	logDir := os.Getenv("HOOK_COST_DIR")
	if logDir == "" {
		logDir = filepath.Join(home, ".cursor", "cost")
	}

	result, code := hooks.CostEstimator(input, logDir)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
