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
		"transcript_path": transcriptPath,
		"file_path":       filePath,
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
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
	}
	if result.Message != "" {
		t.Errorf("expected no message when marker found, got %q", result.Message)
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
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
	}
	if result.Message == "" {
		t.Error("expected warning message when no marker found")
	}
	if !strings.Contains(result.Message, "No self-review detected") {
		t.Errorf("expected message about no self-review, got %q", result.Message)
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
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
	}
	if !strings.Contains(result.Message, "type hints") {
		t.Errorf("expected Python-specific questions, got %q", result.Message)
	}
}

func TestSelfReview_StopHookActive(t *testing.T) {
	result, code := SelfReview(stopInput("", "", true))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
	}
	if result.Message != "" {
		t.Errorf("expected no message when stop_hook_active, got %q", result.Message)
	}
}

func TestSelfReview_NoTranscriptPath(t *testing.T) {
	result, code := SelfReview(stopInput("", "", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
	}
}

func TestSelfReview_MissingTranscript(t *testing.T) {
	result, code := SelfReview(stopInput("/nonexistent/transcript.txt", "", false))
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
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
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %s", result.Decision)
	}
	if result.Message != "" {
		t.Errorf("expected no message when marker found, got %q", result.Message)
	}
}
