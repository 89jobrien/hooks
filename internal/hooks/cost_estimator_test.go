package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCostEstimator_TracksTokens(t *testing.T) {
	dir := t.TempDir()

	CostEstimator(shellInput("ls -la"), dir)
	CostEstimator(writeInput("main.go", "package main\nfunc main() {}"), dir)

	logFile := filepath.Join(dir, "cost.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected cost.log: %v", err)
	}

	log := string(data)
	if !strings.Contains(log, "tokens") {
		t.Errorf("log missing token info, got: %s", log)
	}
}

func TestCostEstimator_AlwaysAllows(t *testing.T) {
	dir := t.TempDir()
	result, code := CostEstimator(shellInput("ls"), dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}

func TestCostEstimator_AccumulatesAcrossCalls(t *testing.T) {
	dir := t.TempDir()

	CostEstimator(shellInput("go test ./..."), dir)
	CostEstimator(shellInput("go build ./..."), dir)

	data, _ := os.ReadFile(filepath.Join(dir, "cost.log"))
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Errorf("expected at least 2 log entries, got %d", len(lines))
	}
}
