package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
)

func main() {
	if hooks.IsHookDisabled("session-guard") {
		fmt.Println(`{}`)
		os.Exit(0)
	}
	data, _ := io.ReadAll(os.Stdin)

	var input hooks.HookInput
	json.Unmarshal(data, &input)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(`{}`)
		os.Exit(0)
	}

	result, code := hooks.SessionGuard(input, cwd)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
