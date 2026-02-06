package hooks

import (
	"encoding/json"
	"testing"
)

func writeInput(path, contents string) HookInput {
	ti, _ := json.Marshal(map[string]string{"path": path, "contents": contents})
	return HookInput{ToolName: "Write", ToolInput: ti}
}

func TestValidateWrite_Blocks(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contents string
	}{
		{".env", ".env", "SECRET=foo"},
		{".env.local", ".env.local", "SECRET=foo"},
		{".env.production", ".env.production", "DB_PASS=bar"},
		{"credentials.json", "config/credentials.json", "{}"},
		{"id_rsa", "~/.ssh/id_rsa", "key"},
		{"id_ed25519", "~/.ssh/id_ed25519", "key"},
		{".pem file", "server.pem", "cert"},
		{".key file", "ssl/server.key", "key"},
		{"secrets.yaml", "k8s/secrets.yaml", "apiVersion: v1"},
		{"secrets.yml", "deploy/secrets.yml", "secret: yes"},
		{".npmrc", ".npmrc", "//registry.npmjs.org/:_authToken=xxx"},
		{".pypirc", "~/.pypirc", "[pypi]\npassword=xxx"},
		{"kubeconfig", "~/.kube/config", "clusters:"},
		{"service account key", "sa-key.json", `{"type": "service_account"}`},
		{".htpasswd", ".htpasswd", "admin:hash"},
		{"terraform.tfvars", "terraform.tfvars", `db_password="secret"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := ValidateWrite(writeInput(tt.path, tt.contents))
			if code != 2 {
				t.Errorf("expected exit 2 (block), got %d for %q", code, tt.path)
			}
			if result.Decision != "deny" {
				t.Errorf("expected deny, got %q", result.Decision)
			}
		})
	}
}

func TestValidateWrite_Allows(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contents string
	}{
		{"Go file", "main.go", "package main"},
		{"Python file", "app.py", `print("hello")`},
		{".env.example", ".env.example", "SECRET=change_me"},
		{".gitignore", ".gitignore", "node_modules/"},
		{"config.yaml", "config.yaml", "port: 8080"},
		{"README.md", "README.md", "# My Project"},
		{"Dockerfile", "Dockerfile", "FROM golang:1.22"},
		{"main.tf", "main.tf", `resource "aws" {}`},
		{"test file", "main_test.go", "package main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := ValidateWrite(writeInput(tt.path, tt.contents))
			if code != 0 {
				t.Errorf("expected exit 0 (allow), got %d for %q", code, tt.path)
			}
			if result.Decision != "allow" {
				t.Errorf("expected allow, got %q", result.Decision)
			}
		})
	}
}

func TestValidateWrite_PassthroughNonWrite(t *testing.T) {
	ti, _ := json.Marshal(map[string]string{"command": "cat .env"})
	input := HookInput{ToolName: "Shell", ToolInput: ti}

	result, code := ValidateWrite(input)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("non-Write tool should passthrough, got code=%d decision=%q", code, result.Decision)
	}
}
