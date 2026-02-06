package hooks

import (
	"os"
	"path/filepath"
	"strings"
)

var selfReviewMarkers = []string{
	"self-review",
	"critical review",
	"implementation complete",
	"testing complete",
	"edge cases considered",
}

// SelfReview is a stop hook that checks if a session included self-review markers.
// It reads the transcript and warns if no markers are found, but always allows (doesn't block).
func SelfReview(input HookInput) (HookResult, int) {
	if input.StopHookActive() {
		return Allow(), 0
	}

	transcriptPath := input.TranscriptPath()
	if transcriptPath == "" {
		return Allow(), 0
	}

	if _, err := os.Stat(transcriptPath); err != nil {
		return Allow(), 0
	}

	content, err := os.ReadFile(transcriptPath)
	if err != nil {
		return Allow(), 0
	}

	contentLower := strings.ToLower(string(content))
	hasMarker := false
	for _, marker := range selfReviewMarkers {
		if strings.Contains(contentLower, marker) {
			hasMarker = true
			break
		}
	}

	if hasMarker {
		return Allow(), 0
	}

	// No marker found - generate warning with review questions
	filePath := input.FilePath()
	questions := generateReviewQuestions(filePath)

	var msg strings.Builder
	msg.WriteString("No self-review detected in session. Consider reviewing:\n")
	for _, q := range questions {
		msg.WriteString("  - ")
		msg.WriteString(q)
		msg.WriteString("\n")
	}

	return AllowMsg(msg.String()), 0
}

func generateReviewQuestions(filePath string) []string {
	questions := []string{
		"Have all requested features been fully implemented?",
		"Are error cases and edge conditions properly handled?",
		"Have you tested the implementation with various inputs?",
		"Is the code documented and readable?",
		"Are there any performance or security concerns?",
	}

	if filePath == "" {
		return questions
	}

	ext := filepath.Ext(filePath)
	switch ext {
	case ".py", ".ts", ".js":
		questions = append(questions,
			"Are type hints/types properly defined?",
			"Is there adequate test coverage?",
		)
	}

	return questions
}
