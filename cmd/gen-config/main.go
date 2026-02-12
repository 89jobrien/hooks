package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"hooks/internal/config"

	"gopkg.in/yaml.v3"
)

var binPrefix = "./hooks/bin/"

func cmd(entry config.HookEntry) string { return binPrefix + entry.Name }

func filterEntries(entries []config.HookEntry) []config.HookEntry {
	var out []config.HookEntry
	for _, e := range entries {
		if e.Included() {
			out = append(out, e)
		}
	}
	return out
}

func allEntries(cfg config.Config) []config.HookEntry {
	var out []config.HookEntry
	out = append(out, filterEntries(cfg.SessionStart)...)
	out = append(out, filterEntries(cfg.BeforeSubmitPrompt)...)
	out = append(out, filterEntries(cfg.PreToolUse)...)
	out = append(out, filterEntries(cfg.PostToolUse)...)
	out = append(out, filterEntries(cfg.Stop)...)
	out = append(out, filterEntries(cfg.PreCompact)...)
	out = append(out, filterEntries(cfg.SessionEnd)...)
	return out
}

func validateHookBinaries(cfg config.Config, binDir string) error {
	seen := make(map[string]bool)
	for _, e := range allEntries(cfg) {
		if e.Name == "" || seen[e.Name] {
			continue
		}
		seen[e.Name] = true
		path := filepath.Join(binDir, e.Name)
		if info, err := os.Stat(path); err != nil {
			return fmt.Errorf("hook %q: binary not found at %s (run: make -C .hooks all or make -C hooks all)", e.Name, path)
		} else if info.IsDir() {
			return fmt.Errorf("hook %q: %s is a directory, expected binary", e.Name, path)
		}
	}
	return nil
}

func main() {
	skipValidate := flag.Bool("skip-validate", false, "skip hook binary existence check (e.g. for init before bins installed)")
	flag.Parse()

	configPath := "hooks/config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = ".hooks/config.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.yaml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read config: %v\n", err)
		os.Exit(1)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "parse config: %v\n", err)
		os.Exit(1)
	}

	// Resolve binPrefix: config.yaml output.binDir > default from config path
	if cfg.Output != nil && cfg.Output.BinDir != "" {
		bp := config.ExpandHome(cfg.Output.BinDir)
		if bp[len(bp)-1] != '/' {
			bp += "/"
		}
		binPrefix = bp
	} else if configPath == ".hooks/config.yaml" {
		binPrefix = "./.hooks/bin/"
	} else if configPath == "config.yaml" {
		binPrefix = "./bin/"
	}
	// else hooks/config.yaml: keep default ./hooks/bin/

	binDir := "bin"
	switch configPath {
	case "hooks/config.yaml":
		binDir = "hooks/bin"
	case ".hooks/config.yaml":
		binDir = ".hooks/bin"
	}
	if !*skipValidate {
		if err := validateHookBinaries(cfg, binDir); err != nil {
			fmt.Fprintf(os.Stderr, "config: %v\n", err)
			os.Exit(1)
		}
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
	openCodeDir := ".opencode"
	if cfg.Output != nil && cfg.Output.OpenCodeDir != "" {
		openCodeDir = cfg.Output.OpenCodeDir
	}
	if d := os.Getenv("HOOK_CONFIG_OPENCODE_DIR"); d != "" {
		openCodeDir = d
	}
	backends := cfg.Output
	if backends == nil {
		backends = &config.Output{}
	}
	wantCursor := wantBackend(backends.Backends, "cursor")
	wantClaude := wantBackend(backends.Backends, "claude")
	wantOpenCode := wantBackend(backends.Backends, "opencode") && openCodeDir != ""

	var cursorJSON []byte
	if wantCursor {
		cursor := cursorConfig(cfg)
		cursorPath := filepath.Join(cursorDir, "hooks.json")
		if err := os.MkdirAll(filepath.Dir(cursorPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
			os.Exit(1)
		}
		cursorJSON, _ = json.MarshalIndent(cursor, "", "  ")
		if err := os.WriteFile(cursorPath, cursorJSON, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", cursorPath, err)
			os.Exit(1)
		}
		fmt.Println("wrote", cursorPath)
	}

	if wantClaude {
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
	}

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

	// Optional: write hooks.json to globalDir (uses same content as Cursor)
	if wantCursor && cfg.Output != nil && cfg.Output.GlobalDir != "" && len(cursorJSON) > 0 {
		globalDir := config.ExpandHome(cfg.Output.GlobalDir)
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

	if wantOpenCode {
		writeOpenCodeOutput(cfg, openCodeDir)
	}
}

// wantBackend returns true if backends is empty (all) or contains name.
func wantBackend(backends []string, name string) bool {
	if len(backends) == 0 {
		return true
	}
	for _, b := range backends {
		if b == name {
			return true
		}
	}
	return false
}

func writeOpenCodeOutput(cfg config.Config, openCodeDir string) {
	manifest := openCodeManifest(cfg)
	pluginsDir := filepath.Join(openCodeDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", pluginsDir, err)
		os.Exit(1)
	}
	manifestPath := filepath.Join(openCodeDir, "hooks-manifest.json")
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", manifestPath, err)
		os.Exit(1)
	}
	fmt.Println("wrote", manifestPath)
	adapterPath := filepath.Join(pluginsDir, "cursor-hooks-adapter.js")
	if err := os.WriteFile(adapterPath, []byte(openCodeAdapterJS), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", adapterPath, err)
		os.Exit(1)
	}
	fmt.Println("wrote", adapterPath)
}

func openCodeManifest(cfg config.Config) map[string]interface{} {
	hookEntries := func(entries []config.HookEntry) []map[string]interface{} {
		out := make([]map[string]interface{}, 0, len(entries))
		for _, e := range filterEntries(entries) {
			m := map[string]interface{}{"command": cmd(e)}
			if e.Matcher != "" {
				m["matcher"] = e.Matcher
			} else {
				m["matcher"] = ".*"
			}
			out = append(out, m)
		}
		return out
	}
	return map[string]interface{}{
		"preToolUse":  hookEntries(cfg.PreToolUse),
		"postToolUse": hookEntries(cfg.PostToolUse),
	}
}

// openCodeAdapterJS is the OpenCode plugin that invokes hook binaries from the manifest.
const openCodeAdapterJS = `// Generated by hooks gen-config. Adapter for Cursor-style hooks (preToolUse/postToolUse).
import path from "path";
import fs from "fs";
import { spawn } from "child_process";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const OPENCODE_DIR = path.join(__dirname, "..");

const OPCODE_TO_CONTRACT = { bash: "Shell", write: "Write", read: "Read", edit: "Edit" };

function loadManifest() {
  const p = path.join(OPENCODE_DIR, "hooks-manifest.json");
  try {
    return JSON.parse(fs.readFileSync(p, "utf8"));
  } catch (e) {
    return null;
  }
}

async function runHooks(manifestKey, toolNameContract, toolInput, directory) {
  const manifest = loadManifest();
  if (!manifest || !manifest[manifestKey] || !Array.isArray(manifest[manifestKey])) return;
  const stdin = JSON.stringify({ tool_name: toolNameContract, tool_input: toolInput }) + "\n";
  for (const hook of manifest[manifestKey]) {
    const matcher = hook.matcher || ".*";
    if (matcher !== ".*" && matcher !== toolNameContract) continue;
    const cmd = hook.command;
    const bin = path.isAbsolute(cmd) ? cmd : path.join(directory, cmd.replace(/^\\.\\//, ""));
    const proc = spawn(bin, [], { stdio: ["pipe", "pipe", "inherit"], cwd: directory, shell: false });
    let out = "";
    proc.stdout.on("data", (chunk) => { out += chunk; });
    proc.stdin.end(stdin);
    await new Promise((resolve, reject) => {
      proc.on("close", (code) => {
        if (code === 2) reject(new Error("Hook blocked (exit 2)"));
        else if (code !== 0) resolve();
        else {
          try {
            const res = JSON.parse(out.split("` + "\\n" + `")[0] || "{}");
            if (res.decision === "deny") reject(new Error(res.reason || "Hook denied"));
            else resolve();
          } catch (e) {
            resolve();
          }
        }
      });
    });
  }
}

export const CursorHooksAdapter = async ({ directory }) => {
  const dir = directory || process.cwd();
  return {
    "tool.execute.before": async (input, output) => {
      const tool = input.tool;
      const contractName = OPCODE_TO_CONTRACT[tool] || tool;
      const toolInput = output.args || {};
      await runHooks("preToolUse", contractName, toolInput, dir);
    },
    "tool.execute.after": async (input, output) => {
      const tool = input.tool;
      const contractName = OPCODE_TO_CONTRACT[tool] || tool;
      const toolInput = output.args || {};
      await runHooks("postToolUse", contractName, toolInput, dir);
    },
  };
};
`

func hasAnyAllowlist(a *config.Allowlists) bool {
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

func buildAllowlistsJSON(a *config.Allowlists) map[string]interface{} {
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

func cursorConfig(cfg config.Config) map[string]interface{} {
	hook := func(entries []config.HookEntry) []map[string]interface{} {
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

func claudeConfig(cfg config.Config) map[string]interface{} {
	hookClause := func(entries []config.HookEntry) []map[string]interface{} {
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

func claudePreToolUse(entries []config.HookEntry) []map[string]interface{} {
	noMatcher := make([]config.HookEntry, 0)
	shell := make([]config.HookEntry, 0)
	write := make([]config.HookEntry, 0)
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

func claudePostToolUse(entries []config.HookEntry) []map[string]interface{} {
	noMatcher := make([]config.HookEntry, 0)
	write := make([]config.HookEntry, 0)
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

func hookList(entries []config.HookEntry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]interface{}{"type": "command", "command": cmd(e)})
	}
	return out
}
