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
	gen-config

BINS := $(addprefix $(BINDIR)/,$(CMDS))

.PHONY: all test clean config summary

all: $(BINS)

# Generate .cursor/hooks.json and .claude/settings.json from hooks/config.yaml.
# Run from repo root: make -C hooks config  (writes to repo root .cursor/ and .claude/).
config: $(BINDIR)/gen-config
	@cd .. && $(CURDIR)/$(BINDIR)/gen-config

# Print audit/cost summary (run from anywhere; uses ~/.cursor/audit and ~/.cursor/cost by default).
summary:
	@bash "$(CURDIR)/scripts/summary.sh"

$(BINDIR)/%: cmd/%/main.go $(wildcard internal/hooks/*.go)
	@mkdir -p $(BINDIR)
	go build -o $@ ./cmd/$*

test:
	go test -v -count=1 ./...

clean:
	rm -rf $(BINDIR)
