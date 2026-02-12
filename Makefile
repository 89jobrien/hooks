BINDIR := bin
CMDS := \
	validate-shell validate-write no-long-running audit lint-on-write lint-changed typecheck-changed shellcheck readonly-guard path-validation session-guard \
	secret-scanner network-fence no-sudo dependency-typosquat \
	test-buddy file-size-guard import-guard check-any-changed todo-tracker \
	branch-guard commit-msg-lint \
	session-diary compact-snapshot prompt-enricher codebase-map jit-context \
	time-tracker-start time-tracker-end \
	rate-limiter cost-estimator dry-run-mode \
	self-review knowledge-update \
	gen-config hooks interactive

BINS := $(addprefix $(BINDIR)/,$(CMDS))

# Install hooks binary to PREFIX/bin (default ~/.local so ~/.local/bin/hooks is on PATH).
PREFIX ?= $(HOME)/.local

.PHONY: all test clean config summary install docker-build docker-run list

all: $(BINS)

install: all
	@mkdir -p $(PREFIX)/bin
	@for f in $(BINS); do install -m 755 $$f $(PREFIX)/bin/$$(basename $$f); done
	@echo "installed $(words $(BINS)) binaries to $(PREFIX)/bin — ensure $(PREFIX)/bin is on your PATH"

# Generate .cursor/hooks.json and .claude/settings.json. Writes to cwd (repo root).
# From this repo: make config. From a repo that has hooks as subdir: make -C hooks config.
# From a repo with install.sh layout: run ./.hooks/bin/gen-config (no Makefile in .hooks).
config: $(BINDIR)/gen-config
	@if [ ".hooks" = "$(notdir $(CURDIR))" ]; then cd .. && $(CURDIR)/$(BINDIR)/gen-config; \
	elif [ "hooks" = "$(notdir $(CURDIR))" ]; then cd .. && $(CURDIR)/$(BINDIR)/gen-config; \
	else $(CURDIR)/$(BINDIR)/gen-config; fi

# Print audit/cost summary (run from anywhere; uses ~/.cursor/audit and ~/.cursor/cost by default).
summary:
	@bash "$(CURDIR)/scripts/summary.sh"

$(BINDIR)/%: cmd/%/main.go $(wildcard internal/hooks/*.go) $(wildcard internal/config/*.go)
	@mkdir -p $(BINDIR)
	go build -o $@ ./cmd/$*

test:
	go test -v -count=1 ./...

clean:
	rm -rf $(BINDIR)

docker-build:
	docker build -t ghcr.io/89jobrien/hooks:local .

docker-run:
	docker run --rm ghcr.io/89jobrien/hooks:local audit

list:
	@echo "Available targets:"
	@echo "  all          - Build all hook binaries"
	@echo "  install      - Install binaries to PREFIX/bin (default ~/.local/bin)"
	@echo "  config       - Generate .cursor/hooks.json and .claude/settings.json"
	@echo "  summary      - Print audit/cost summary"
	@echo "  test         - Run Go tests"
	@echo "  clean        - Remove bin/ directory"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  list         - Show this help"
