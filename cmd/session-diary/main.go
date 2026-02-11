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
	if hooks.IsHookDisabled("session-diary") {
		fmt.Println(`{}`)
		os.Exit(0)
	}
	data, _ := io.ReadAll(os.Stdin)
	var input hooks.HookInput
	json.Unmarshal(data, &input)

	home, _ := os.UserHomeDir()
	auditDir := os.Getenv("HOOK_AUDIT_DIR")
	if auditDir == "" {
		auditDir = filepath.Join(home, ".cursor", "audit")
	}
	diaryDir := os.Getenv("HOOK_DIARY_DIR")
	if diaryDir == "" {
		diaryDir = filepath.Join(home, ".cursor", "diary")
	}

	result, code := hooks.SessionDiary(input, auditDir, diaryDir)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
