package hooks

import (
	"strings"
	"testing"
)

func join(parts ...string) string {
	return strings.Join(parts, "")
}

func TestSecretScanner_Detects(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contents string
	}{
		{"AWS access key", "config.go", "var key = \"" + join("AKIA", "IOSFODNN7EXAMPLE") + "\""},
		{"AWS secret key", "config.go", "aws_secret_access_key = \"" + join("wJalrXUtnFEMI/K7MDENG/bPxRfi", "CYEXAMPLEKEY") + "\""},
		{"GitHub token ghp_", "main.go", "token := \"" + join("ghp_", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef12") + "\""},
		{"GitHub token github_pat_", "main.go", "token := \"" + join("github_pat_", "ABCDEFGHIJKLMNOP_", "abcdefghijklmnop") + "\""},
		{"Slack token xoxb-", "bot.py", "SLACK_TOKEN = \"" + join("xoxb-", "123456789-", "123456789-", "abcdefghijklmnop") + "\""},
		{"Slack token xoxp-", "bot.py", "SLACK_TOKEN = \"" + join("xoxp-", "123456789-", "123456789-", "abcdefghijklmnop") + "\""},
		{"generic password=", "config.yaml", "password: \"" + join("s3cret", "P@ssw0rd!") + "\""},
		{"generic secret=", "app.go", "var secret = \"" + join("my-super-secret", "-value-1234") + "\""},
		{"private key header", "cert.go", "key := \"" + join("-----BEGIN RSA PRIVATE KEY-----", "\\nMIIE...") + "\""},
		{"generic api_key=", "config.py", "API_KEY = \"" + join("sk-", "abcdefghijklmnopqrstuvwxyz123456") + "\""},
		{"Stripe key sk_live", "billing.go", "var stripeKey = \"" + join("sk_live_", "abcdefghijklmnopqrstuvwxyz") + "\""},
		{"Stripe key sk_test", "billing.go", "var stripeKey = \"" + join("sk_test_", "abcdefghijklmnopqrstuvwxyz") + "\""},
		{"Heroku API key", "deploy.sh", "HEROKU_API_KEY=\"" + join("abcdef12-", "3456-7890-abcd-ef1234567890") + "\""},
		{"JWT token", "auth.go", "token := \"" + join("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.", "eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.", "SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c") + "\""},
		{"SendGrid key SG.", "email.go", "apiKey := \"" + join("SG.", "abcdefghijklmnop.", "abcdefghijklmnopqrstuvwxyz123456") + "\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := SecretScanner(writeInput(tt.path, tt.contents))
			if code != 2 {
				t.Errorf("expected block (exit 2) for %s, got %d", tt.name, code)
			}
			if result.Decision != "deny" {
				t.Errorf("expected deny, got %q", result.Decision)
			}
		})
	}
}

func TestSecretScanner_AllowsClean(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contents string
	}{
		{"normal Go code", "main.go", `package main\nfunc main() { fmt.Println("hello") }`},
		{"normal Python", "app.py", `def handler(request):\n    return Response(200)`},
		{"config with placeholder", "config.yaml", `api_key: "${API_KEY}"\npassword: "${DB_PASSWORD}"`},
		{"env var reference", "main.go", `key := os.Getenv("API_KEY")`},
		{"test with fake key", "main_test.go", `key := "test-key-not-real"`},
		{".env.example", ".env.example", `API_KEY=your-key-here\nSECRET=change-me`},
		{"base64 short string", "util.go", `encoded := "aGVsbG8="`},
		{"password field name", "model.go", `type User struct { Password string }`},
		{"non-Write tool", "main.go", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input HookInput
			if tt.name == "non-Write tool" {
				input = shellInput("ls")
			} else {
				input = writeInput(tt.path, tt.contents)
			}
			result, code := SecretScanner(input)
			if code != 0 {
				t.Errorf("expected allow (exit 0) for %s, got %d; reason: %s", tt.name, code, result.Reason)
			}
		})
	}
}
