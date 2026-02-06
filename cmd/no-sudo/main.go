package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
	"strings"
)

func main() {
	if hooks.IsHookDisabled("no-sudo") {
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

	if !optIn("HOOK_NO_SUDO") {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}

	result, code := hooks.NoSudo(input)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}

func optIn(name string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	return v == "1" || v == "true" || v == "yes"
}
