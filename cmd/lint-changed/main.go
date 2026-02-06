package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
)

func main() {
	if hooks.IsHookDisabled("lint-changed") {
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

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}

	result, code := hooks.LintChanged(input, cwd)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
