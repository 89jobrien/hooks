package hooks

import (
	"testing"
)

func TestImportGuard_BlocksBannedImports(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contents string
	}{
		{"Go fmt.Println in non-main", "handler.go", "package handler\nimport \"fmt\"\nfunc Handle() { fmt.Println(\"debug\") }"},
		{"Go os/exec in handler", "api/handler.go", "package api\nimport \"os/exec\"\nfunc Run() { exec.Command(\"ls\") }"},
		{"Go reflect in model", "model.go", "package model\nimport \"reflect\"\nvar t = reflect.TypeOf(0)"},
		{"Python os.system", "app.py", "import os\nos.system('ls')"},
		{"Python eval", "util.py", "result = eval(user_input)"},
		{"Python exec", "run.py", "exec(code_string)"},
		{"JS eval", "app.js", "const result = eval(userCode);"},
	}

	banned := map[string][]string{
		".go": {"os/exec", "reflect", "fmt.Println"},
		".py": {"os.system", "eval(", "exec("},
		".js": {"eval("},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := ImportGuard(writeInput(tt.path, tt.contents), banned)
			if code != 2 {
				t.Errorf("expected block (exit 2), got %d for %s", code, tt.name)
			}
			if result.Decision != "deny" {
				t.Errorf("expected deny, got %q", result.Decision)
			}
		})
	}
}

func TestImportGuard_AllowsClean(t *testing.T) {
	banned := map[string][]string{
		".go": {"os/exec", "reflect", "fmt.Println"},
		".py": {"os.system", "eval(", "exec("},
		".js": {"eval("},
	}

	tests := []struct {
		name     string
		path     string
		contents string
	}{
		{"Go clean code", "handler.go", "package handler\nimport \"net/http\"\nfunc Handle(w http.ResponseWriter) {}"},
		{"Python clean", "app.py", "from flask import Flask\napp = Flask(__name__)"},
		{"JS clean", "app.js", "const express = require('express');\nconst app = express();"},
		{"Go test with fmt", "handler_test.go", "package handler\nimport \"fmt\"\nfunc TestX() { fmt.Println(\"ok\") }"},
		{"unknown extension", "data.csv", "a,b,c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := ImportGuard(writeInput(tt.path, tt.contents), banned)
			if code != 0 {
				t.Errorf("expected allow (exit 0), got %d for %s; reason: %s", code, tt.name, result.Reason)
			}
		})
	}
}

func TestImportGuard_PassthroughNonWrite(t *testing.T) {
	result, code := ImportGuard(shellInput("ls"), nil)
	if code != 0 || result.Decision != "allow" {
		t.Error("should passthrough non-Write tools")
	}
}

func TestImportGuardWithAllowlist_AllowsListedPattern(t *testing.T) {
	banned := map[string][]string{".go": {"os/exec", "reflect", "fmt.Println"}}
	allowed := map[string][]string{".go": {"os/exec"}} // allow os/exec in .go
	result, code := ImportGuardWithAllowlist(writeInput("handler.go", "package p\nimport \"os/exec\"\nfunc Run() { exec.Command(\"ls\") }"), banned, allowed)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when os/exec in allowlist for .go, got code=%d decision=%q", code, result.Decision)
	}
}

func TestImportGuardWithAllowlist_BlocksWhenNotInAllowlist(t *testing.T) {
	banned := map[string][]string{".go": {"os/exec"}}
	allowed := map[string][]string{".go": {"reflect"}} // only reflect allowed, not os/exec
	result, code := ImportGuardWithAllowlist(writeInput("x.go", "package p\nimport \"os/exec\"\nfunc Run() {}"), banned, allowed)
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny when os/exec not in allowlist, got code=%d decision=%q", code, result.Decision)
	}
}
