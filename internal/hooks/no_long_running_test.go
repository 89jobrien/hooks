package hooks

import (
	"testing"
)

func TestNoLongRunning_Blocks(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"npm run dev", "npm run dev"},
		{"npm start", "npm start"},
		{"yarn dev", "yarn dev"},
		{"pnpm dev", "pnpm dev"},
		{"npx next dev", "npx next dev"},
		{"python http.server", "python -m http.server"},
		{"python3 http.server", "python3 -m http.server 8080"},
		{"flask run", "flask run"},
		{"uvicorn", "uvicorn app:app --reload"},
		{"gunicorn", "gunicorn -w 4 app:app"},
		{"go run server", "go run ./cmd/server"},
		{"air", "air"},
		{"cargo watch", "cargo watch -x run"},
		{"docker compose up", "docker compose up"},
		{"docker-compose up", "docker-compose up"},
		{"fswatch", "fswatch -o . | xargs -n1 make"},
		{"inotifywait", "inotifywait -m -r ."},
		{"tail -f", "tail -f /var/log/syslog"},
		{"npm run watch", "npm run watch"},
		{"npm run serve", "npm run serve"},
		{"nodemon", "nodemon server.js"},
		{"hugo server", "hugo server"},
		{"jekyll serve", "jekyll serve"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := NoLongRunning(shellInput(tt.cmd))
			if code != 2 {
				t.Errorf("expected exit 2 (block), got %d for %q", code, tt.cmd)
			}
			if result.Decision != "deny" {
				t.Errorf("expected deny, got %q", result.Decision)
			}
		})
	}
}

func TestNoLongRunning_Allows(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"npm install", "npm install"},
		{"npm test", "npm test"},
		{"npm run build", "npm run build"},
		{"npm run lint", "npm run lint"},
		{"go test", "go test ./..."},
		{"go build", "go build ./..."},
		{"go run (not server)", "go run main.go"},
		{"cargo build", "cargo build"},
		{"cargo test", "cargo test"},
		{"python script", "python3 migrate.py"},
		{"pytest", "pytest -v"},
		{"docker build", "docker build -t app ."},
		{"docker compose up -d", "docker compose up -d"},
		{"make", "make build"},
		{"git log", "git log --oneline -10"},
		{"tail -n", "tail -n 50 output.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := NoLongRunning(shellInput(tt.cmd))
			if code != 0 {
				t.Errorf("expected exit 0 (allow), got %d for %q", code, tt.cmd)
			}
			if result.Decision != "allow" {
				t.Errorf("expected allow, got %q", result.Decision)
			}
		})
	}
}

func TestNoLongRunning_PassthroughNonShell(t *testing.T) {
	result, code := NoLongRunning(writeInput("main.go", "package main"))
	if code != 0 || result.Decision != "allow" {
		t.Errorf("non-Shell tool should passthrough")
	}
}
