package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type HookEntry struct {
	Name    string `yaml:"name"`
	Matcher string `yaml:"matcher,omitempty"`
	Enabled *bool  `yaml:"enabled,omitempty"` // nil or true = include; false = omit from output
}

func (h *HookEntry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		h.Name = s
		return nil
	}
	var m struct {
		Name    string `yaml:"name"`
		Matcher string `yaml:"matcher"`
		Enabled *bool  `yaml:"enabled"`
	}
	if err := unmarshal(&m); err != nil {
		return err
	}
	h.Name = m.Name
	h.Matcher = m.Matcher
	h.Enabled = m.Enabled
	return nil
}

type Allowlists struct {
	NetworkFence *struct {
		AllowedDomains []string `yaml:"allowedDomains"`
	} `yaml:"networkFence,omitempty"`
	DependencyTyposquat *struct {
		AllowedPackages []string `yaml:"allowedPackages"`
	} `yaml:"dependencyTyposquat,omitempty"`
	ImportGuard *struct {
		AllowedPatterns map[string][]string `yaml:"allowedPatterns"` // ext -> patterns to allow
	} `yaml:"importGuard,omitempty"`
}

type Output struct {
	BinDir    string `yaml:"binDir,omitempty"`
	CursorDir string `yaml:"cursorDir,omitempty"`
	ClaudeDir string `yaml:"claudeDir,omitempty"`
	GlobalDir string `yaml:"globalDir,omitempty"`
}

type Config struct {
	Version            int               `yaml:"version"`
	Env                map[string]string `yaml:"env,omitempty"`
	Output             *Output           `yaml:"output,omitempty"`
	Allowlists         *Allowlists       `yaml:"allowlists,omitempty"`
	SessionStart       []HookEntry       `yaml:"sessionStart"`
	BeforeSubmitPrompt []HookEntry       `yaml:"beforeSubmitPrompt"`
	PreToolUse         []HookEntry       `yaml:"preToolUse"`
	PostToolUse        []HookEntry       `yaml:"postToolUse"`
	Stop               []HookEntry       `yaml:"stop"`
	PreCompact         []HookEntry       `yaml:"preCompact"`
	SessionEnd         []HookEntry       `yaml:"sessionEnd"`
}

var binPrefix = "./hooks/bin/"

func cmd(entry HookEntry) string { return binPrefix + entry.Name }

func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, path[1:])
}

func (h HookEntry) included() bool {
	return h.Enabled == nil || *h.Enabled
}

func filterEntries(entries []HookEntry) []HookEntry {
	var out []HookEntry
	for _, e := range entries {
		if e.included() {
			out = append(out, e)
		}
	}
	return out
}

func allEntries(cfg Config) []HookEntry {
	var out []HookEntry
	out = append(out, filterEntries(cfg.SessionStart)...)
	out = append(out, filterEntries(cfg.BeforeSubmitPrompt)...)
	out = append(out, filterEntries(cfg.PreToolUse)...)
	out = append(out, filterEntries(cfg.PostToolUse)...)
	out = append(out, filterEntries(cfg.Stop)...)
	out = append(out, filterEntries(cfg.PreCompact)...)
	out = append(out, filterEntries(cfg.SessionEnd)...)
	return out
}

func validateHookBinaries(cfg Config, binDir string) error {
	seen := make(map[string]bool)
	for _, e := range allEntries(cfg) {
		if e.Name == "" || seen[e.Name] {
			continue
		}
		seen[e.Name] = true
		path := filepath.Join(binDir, e.Name)
		if info, err := os.Stat(path); err != nil {
			return fmt.Errorf("hook %q: binary not found at %s (run: make -C hooks all)", e.Name, path)
		} else if info.IsDir() {
			return fmt.Errorf("hook %q: %s is a directory, expected binary", e.Name, path)
		}
	}
	return nil
}

func main() {
	configPath := "hooks/config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.yaml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read config: %v\n", err)
		os.Exit(1)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "parse config: %v\n", err)
		os.Exit(1)
	}

	// Resolve binPrefix: config.yaml output.binDir > default
	if cfg.Output != nil && cfg.Output.BinDir != "" {
		bp := cfg.Output.BinDir
		if bp[len(bp)-1] != '/' {
			bp += "/"
		}
		binPrefix = bp
	}

	binDir := "bin"
	if configPath == "hooks/config.yaml" {
		binDir = "hooks/bin"
	}
	if err := validateHookBinaries(cfg, binDir); err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	// Resolve output dirs: env var > config.yaml > defaults
	cursorDir := ".cursor"
	claudeDir := ".claude"
	if cfg.Output != nil && cfg.Output.CursorDir != "" {
		cursorDir = cfg.Output.CursorDir
	}
	if cfg.Output != nil && cfg.Output.ClaudeDir != "" {
		claudeDir = cfg.Output.ClaudeDir
	}
	if d := os.Getenv("HOOK_CONFIG_CURSOR_DIR"); d != "" {
		cursorDir = d
	}
	if d := os.Getenv("HOOK_CONFIG_CLAUDE_DIR"); d != "" {
		claudeDir = d
	}

	// Cursor .cursor/hooks.json
	cursor := cursorConfig(cfg)
	cursorPath := filepath.Join(cursorDir, "hooks.json")
	if err := os.MkdirAll(filepath.Dir(cursorPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	cursorJSON, _ := json.MarshalIndent(cursor, "", "  ")
	if err := os.WriteFile(cursorPath, cursorJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", cursorPath, err)
		os.Exit(1)
	}
	fmt.Println("wrote", cursorPath)

	// Claude .claude/settings.json
	claude := claudeConfig(cfg)
	claudePath := filepath.Join(claudeDir, "settings.json")
	if err := os.MkdirAll(filepath.Dir(claudePath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	claudeJSON, _ := json.MarshalIndent(claude, "", "  ")
	if err := os.WriteFile(claudePath, claudeJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", claudePath, err)
		os.Exit(1)
	}
	fmt.Println("wrote", claudePath)

	// Optional .cursor/hooks.env from config.env
	if len(cfg.Env) > 0 {
		envPath := filepath.Join(cursorDir, "hooks.env")
		keys := make([]string, 0, len(cfg.Env))
		for k := range cfg.Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var sb []byte
		for _, k := range keys {
			sb = append(sb, []byte(k+"="+cfg.Env[k]+"\n")...)
		}
		if err := os.WriteFile(envPath, sb, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", envPath, err)
			os.Exit(1)
		}
		fmt.Println("wrote", envPath)
	}

	// Optional .cursor/hooks-allowlists.json from config.allowlists
	if cfg.Allowlists != nil && hasAnyAllowlist(cfg.Allowlists) {
		allowPath := filepath.Join(cursorDir, "hooks-allowlists.json")
		allowJSON := buildAllowlistsJSON(cfg.Allowlists)
		data, _ := json.MarshalIndent(allowJSON, "", "  ")
		if err := os.WriteFile(allowPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", allowPath, err)
			os.Exit(1)
		}
		fmt.Println("wrote", allowPath)
	}

	// Optional: write hooks.json to globalDir
	if cfg.Output != nil && cfg.Output.GlobalDir != "" {
		globalDir := expandHome(cfg.Output.GlobalDir)
		globalPath := filepath.Join(globalDir, "hooks.json")
		if err := os.MkdirAll(globalDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(globalPath, cursorJSON, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", globalPath, err)
			os.Exit(1)
		}
		fmt.Println("wrote", globalPath)
	}
}

func hasAnyAllowlist(a *Allowlists) bool {
	if a == nil {
		return false
	}
	if a.NetworkFence != nil && len(a.NetworkFence.AllowedDomains) > 0 {
		return true
	}
	if a.DependencyTyposquat != nil && len(a.DependencyTyposquat.AllowedPackages) > 0 {
		return true
	}
	if a.ImportGuard != nil && len(a.ImportGuard.AllowedPatterns) > 0 {
		return true
	}
	return false
}

func buildAllowlistsJSON(a *Allowlists) map[string]interface{} {
	out := make(map[string]interface{})
	if a.NetworkFence != nil && len(a.NetworkFence.AllowedDomains) > 0 {
		out["networkFence"] = map[string]interface{}{"allowedDomains": a.NetworkFence.AllowedDomains}
	}
	if a.DependencyTyposquat != nil && len(a.DependencyTyposquat.AllowedPackages) > 0 {
		out["dependencyTyposquat"] = map[string]interface{}{"allowedPackages": a.DependencyTyposquat.AllowedPackages}
	}
	if a.ImportGuard != nil && len(a.ImportGuard.AllowedPatterns) > 0 {
		out["importGuard"] = map[string]interface{}{"allowedPatterns": a.ImportGuard.AllowedPatterns}
	}
	return out
}

func cursorConfig(cfg Config) map[string]interface{} {
	hook := func(entries []HookEntry) []map[string]interface{} {
		out := make([]map[string]interface{}, 0, len(entries))
		for _, e := range entries {
			m := map[string]interface{}{"command": cmd(e)}
			if e.Matcher != "" {
				m["matcher"] = e.Matcher
			}
			out = append(out, m)
		}
		return out
	}
	return map[string]interface{}{
		"version": cfg.Version,
		"hooks": map[string]interface{}{
			"sessionStart":       hook(filterEntries(cfg.SessionStart)),
			"beforeSubmitPrompt": hook(filterEntries(cfg.BeforeSubmitPrompt)),
			"preToolUse":         hook(filterEntries(cfg.PreToolUse)),
			"postToolUse":        hook(filterEntries(cfg.PostToolUse)),
			"stop":               hook(filterEntries(cfg.Stop)),
			"preCompact":         hook(filterEntries(cfg.PreCompact)),
			"sessionEnd":         hook(filterEntries(cfg.SessionEnd)),
		},
	}
}

func claudeConfig(cfg Config) map[string]interface{} {
	hookClause := func(entries []HookEntry) []map[string]interface{} {
		out := make([]map[string]interface{}, 0, len(entries))
		for _, e := range entries {
			out = append(out, map[string]interface{}{"type": "command", "command": cmd(e)})
		}
		return out
	}
	return map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart":     []map[string]interface{}{{"matcher": ".*", "hooks": hookClause(filterEntries(cfg.SessionStart))}},
			"UserPromptSubmit": []map[string]interface{}{{"matcher": ".*", "hooks": hookClause(filterEntries(cfg.BeforeSubmitPrompt))}},
			"PreToolUse":       claudePreToolUse(filterEntries(cfg.PreToolUse)),
			"PostToolUse":      claudePostToolUse(filterEntries(cfg.PostToolUse)),
			"Stop":             []map[string]interface{}{{"matcher": ".*", "hooks": hookClause(filterEntries(cfg.Stop))}},
			"PreCompact":       []map[string]interface{}{{"matcher": ".*", "hooks": hookClause(filterEntries(cfg.PreCompact))}},
			"SessionEnd":       []map[string]interface{}{{"matcher": ".*", "hooks": hookClause(filterEntries(cfg.SessionEnd))}},
		},
	}
}

func claudePreToolUse(entries []HookEntry) []map[string]interface{} {
	noMatcher := make([]HookEntry, 0)
	shell := make([]HookEntry, 0)
	write := make([]HookEntry, 0)
	for _, e := range entries {
		switch e.Matcher {
		case "":
			noMatcher = append(noMatcher, e)
		case "Shell":
			shell = append(shell, e)
		case "Write":
			write = append(write, e)
		default:
			noMatcher = append(noMatcher, e)
		}
	}
	var out []map[string]interface{}
	if len(noMatcher) > 0 {
		out = append(out, map[string]interface{}{"matcher": ".*", "hooks": hookList(noMatcher)})
	}
	if len(shell) > 0 {
		out = append(out, map[string]interface{}{"matcher": "Shell", "hooks": hookList(shell)})
	}
	if len(write) > 0 {
		out = append(out, map[string]interface{}{"matcher": "Write", "hooks": hookList(write)})
	}
	return out
}

func claudePostToolUse(entries []HookEntry) []map[string]interface{} {
	noMatcher := make([]HookEntry, 0)
	write := make([]HookEntry, 0)
	for _, e := range entries {
		if e.Matcher == "Write" {
			write = append(write, e)
		} else {
			noMatcher = append(noMatcher, e)
		}
	}
	var out []map[string]interface{}
	if len(noMatcher) > 0 {
		out = append(out, map[string]interface{}{"matcher": ".*", "hooks": hookList(noMatcher)})
	}
	if len(write) > 0 {
		out = append(out, map[string]interface{}{"matcher": "Write", "hooks": hookList(write)})
	}
	return out
}

func hookList(entries []HookEntry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]interface{}{"type": "command", "command": cmd(e)})
	}
	return out
}
