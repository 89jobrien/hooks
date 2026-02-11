package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// HookInput is the JSON payload piped to hooks via stdin.
type HookInput struct {
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

// Command extracts the "command" field from tool_input (Shell tool).
func (h *HookInput) Command() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["command"].(string); ok {
		return v
	}
	return ""
}

// Path extracts the "path" field from tool_input (Write/Read tool).
func (h *HookInput) Path() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["path"].(string); ok {
		return v
	}
	return ""
}

// Contents extracts the "contents" field from tool_input (Write tool).
func (h *HookInput) Contents() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["contents"].(string); ok {
		return v
	}
	return ""
}

// Pattern extracts the "pattern" field from tool_input (Grep tool).
func (h *HookInput) Pattern() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["pattern"].(string); ok {
		return v
	}
	return ""
}

// HookResult is the JSON output from a hook.
type HookResult struct {
	Decision    string `json:"decision,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Message     string `json:"message,omitempty"`
	LintCommand string `json:"lint_command,omitempty"`
}

func Allow() HookResult {
	return HookResult{Decision: "allow"}
}

func AllowMsg(msg string) HookResult {
	return HookResult{Decision: "allow", Message: msg}
}

// NoOp returns an empty result for hooks that don't support the decision field
// (SessionStart, SessionEnd, Stop, PreCompact).
func NoOp() HookResult {
	return HookResult{}
}

// NoOpMsg returns a result with only a reason for session/lifecycle hooks.
func NoOpMsg(msg string) HookResult {
	return HookResult{Reason: msg}
}

func Deny(reason string) HookResult {
	return HookResult{Decision: "deny", Reason: reason}
}

// Prompt extracts the "prompt" field from tool_input (beforeSubmitPrompt).
func (h *HookInput) Prompt() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["prompt"].(string); ok {
		return v
	}
	return ""
}

// TranscriptPath extracts the "transcript_path" field from tool_input (Stop event).
func (h *HookInput) TranscriptPath() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["transcript_path"].(string); ok {
		return v
	}
	return ""
}

// FilePath extracts the "file_path" field from tool_input (Stop event).
func (h *HookInput) FilePath() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["file_path"].(string); ok {
		return v
	}
	return ""
}

// StopHookActive checks if "stop_hook_active" field is true in tool_input (Stop event).
func (h *HookInput) StopHookActive() bool {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return false
	}
	if v, ok := m["stop_hook_active"].(bool); ok {
		return v
	}
	return false
}

// SessionID extracts the "session_id" field from tool_input (SessionStart/beforeSubmitPrompt events).
func (h *HookInput) SessionID() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["session_id"].(string); ok {
		return v
	}
	return ""
}

// Cwd extracts the "cwd" field from tool_input (SessionStart/beforeSubmitPrompt events).
func (h *HookInput) Cwd() string {
	var m map[string]interface{}
	if err := json.Unmarshal(h.ToolInput, &m); err != nil {
		return ""
	}
	if v, ok := m["cwd"].(string); ok {
		return v
	}
	return ""
}

// ReadInput reads and parses HookInput from the given reader.
func ReadInput(r io.Reader) (HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return HookInput{}, fmt.Errorf("reading stdin: %w", err)
	}
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return HookInput{}, fmt.Errorf("parsing input: %w", err)
	}
	return input, nil
}

// IsHookDisabled returns true if name is listed in HOOK_DISABLED (comma-separated, trimmed).
func IsHookDisabled(name string) bool {
	v := os.Getenv("HOOK_DISABLED")
	if v == "" {
		return false
	}
	for _, s := range strings.Split(v, ",") {
		if strings.TrimSpace(s) == name {
			return true
		}
	}
	return false
}

// Run is the standard entrypoint for a hook binary.
// It reads stdin, calls the hook function, writes the JSON result to stdout,
// and exits with the appropriate code.
func Run(hookFn func(HookInput) (HookResult, int)) {
	input, err := ReadInput(os.Stdin)
	if err != nil {
		// Fail open on parse errors
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}

	result, exitCode := hookFn(input)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(exitCode)
}

// RunOrDisabled runs the hook unless its name is in HOOK_DISABLED; then outputs allow and exits 0.
func RunOrDisabled(name string, hookFn func(HookInput) (HookResult, int)) {
	if IsHookDisabled(name) {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}
	Run(hookFn)
}
