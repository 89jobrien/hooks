package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
)

func main() {
	if hooks.IsHookDisabled("self-review") {
		fmt.Println(`{}`)
		os.Exit(0)
	}
	data, _ := io.ReadAll(os.Stdin)
	var input hooks.HookInput
	json.Unmarshal(data, &input)

	result, code := hooks.SelfReview(input)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
