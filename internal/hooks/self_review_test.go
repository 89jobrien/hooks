package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stopInput(transcriptPath, filePath string, stopHookActive bool) HookInput {
	m := map[string]interface{}{
		"transcript_path":  transcriptPath,
		"file_path":        filePath,
		"stop_hook_active": stopHookActive,
	}
	ti, _ := json.Marshal(m)
	return HookInput{ToolName: "Stop", ToolInput: ti}
}

func TestSelfReview_HasMarker(t *testing.T) {
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.txt")
	os.WriteFile(transcript, []byte("Implementation complete. Self-review done."), 0644)

	result, code := SelfReview(stopInput(transcript, "", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for stop hook, got %s", result.Decision)
	}
	if result.Reason != "" {
		t.Errorf("expected no reason when marker found, got %q", result.Reason)
	}
}

func TestSelfReview_NoMarker(t *testing.T) {
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.txt")
	os.WriteFile(transcript, []byte("Just some regular work here."), 0644)

	result, code := SelfReview(stopInput(transcript, "", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Reason == "" {
		t.Error("expected warning reason when no marker found")
	}
	if !strings.Contains(result.Reason, "No self-review detected") {
		t.Errorf("expected reason about no self-review, got %q", result.Reason)
	}
}

func TestSelfReview_NoMarkerWithPythonFile(t *testing.T) {
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.txt")
	os.WriteFile(transcript, []byte("Just some regular work here."), 0644)

	result, code := SelfReview(stopInput(transcript, "test.py", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(result.Reason, "type hints") {
		t.Errorf("expected Python-specific questions, got %q", result.Reason)
	}
}

func TestSelfReview_StopHookActive(t *testing.T) {
	result, code := SelfReview(stopInput("", "", true))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for stop hook, got %s", result.Decision)
	}
}

func TestSelfReview_NoTranscriptPath(t *testing.T) {
	result, code := SelfReview(stopInput("", "", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for stop hook, got %s", result.Decision)
	}
}

func TestSelfReview_MissingTranscript(t *testing.T) {
	result, code := SelfReview(stopInput("/nonexistent/transcript.txt", "", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for stop hook, got %s", result.Decision)
	}
}

func TestSelfReview_MultipleMarkers(t *testing.T) {
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.txt")
	os.WriteFile(transcript, []byte("Testing complete. Edge cases considered."), 0644)

	result, code := SelfReview(stopInput(transcript, "", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for stop hook, got %s", result.Decision)
	}
	if result.Reason != "" {
		t.Errorf("expected no reason when marker found, got %q", result.Reason)
	}
}
