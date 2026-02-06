package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
)

func main() {
	if hooks.IsHookDisabled("prompt-enricher") {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}
	data, _ := io.ReadAll(os.Stdin)
	var input hooks.HookInput
	json.Unmarshal(data, &input)

	cwd, _ := os.Getwd()
	result, code := hooks.PromptEnricher(input, cwd)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
