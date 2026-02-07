package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type HookEntry struct {
	Name    string `yaml:"name"`
	Matcher string `yaml:"matcher,omitempty"`
	Enabled *bool  `yaml:"enabled,omitempty"`
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

func (h HookEntry) Included() bool {
	return h.Enabled == nil || *h.Enabled
}

type Allowlists struct {
	NetworkFence *struct {
		AllowedDomains []string `yaml:"allowedDomains"`
	} `yaml:"networkFence,omitempty"`
	DependencyTyposquat *struct {
		AllowedPackages []string `yaml:"allowedPackages"`
	} `yaml:"dependencyTyposquat,omitempty"`
	ImportGuard *struct {
		AllowedPatterns map[string][]string `yaml:"allowedPatterns"`
	} `yaml:"importGuard,omitempty"`
}

type Config struct {
	Version            int               `yaml:"version"`
	Env                map[string]string `yaml:"env,omitempty"`
	Allowlists         *Allowlists       `yaml:"allowlists,omitempty"`
	SessionStart       []HookEntry       `yaml:"sessionStart"`
	BeforeSubmitPrompt []HookEntry       `yaml:"beforeSubmitPrompt"`
	PreToolUse         []HookEntry       `yaml:"preToolUse"`
	PostToolUse        []HookEntry       `yaml:"postToolUse"`
	Stop               []HookEntry       `yaml:"stop"`
	PreCompact         []HookEntry       `yaml:"preCompact"`
	SessionEnd         []HookEntry       `yaml:"sessionEnd"`
}

// EventName and entries for that event.
type EventEntries struct {
	Event   string
	Entries *[]HookEntry
}

func (c *Config) Events() []EventEntries {
	return []EventEntries{
		{"sessionStart", &c.SessionStart},
		{"beforeSubmitPrompt", &c.BeforeSubmitPrompt},
		{"preToolUse", &c.PreToolUse},
		{"postToolUse", &c.PostToolUse},
		{"stop", &c.Stop},
		{"preCompact", &c.PreCompact},
		{"sessionEnd", &c.SessionEnd},
	}
}

// FindConfigPath returns path to config.yaml (hooks/config.yaml or config.yaml) and work dir (repo root or hooks dir).
// FindConfigPath searches upward from the current working directory for a configuration file and returns the file path and the directory that contains it.
// It looks for "hooks/config.yaml" first and then "config.yaml" in each directory. If no config is found, it returns an error describing the directory where the search started.
func FindConfigPath() (configPath, workDir string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	startDir := dir
	for {
		p := filepath.Join(dir, "hooks", "config.yaml")
		if _, err := os.Stat(p); err == nil {
			return p, dir, nil
		}
		p = filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(p); err == nil {
			return p, dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("no hooks/config.yaml or config.yaml found (searched up from %s)", startDir)
		}
		dir = parent
	}
}

// GlobalHooksPath returns the path to the global hooks configuration file (~/.cursor/hooks.json).
func GlobalHooksPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "hooks.json")
}

// Load reads a YAML configuration file from the given path and unmarshals it into a Config.
// It returns the parsed Config or an error if the file cannot be read or the YAML cannot be parsed.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save marshals cfg to YAML and writes it to the given path.
// It returns an error if marshaling or writing the file fails.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}