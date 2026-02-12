package main

import (
	"encoding/json"
	"testing"

	"hooks/internal/config"
)

func TestBuildAllowlistsJSON_EmitsDependencyTyposquatAndImportGuard(t *testing.T) {
	a := &config.Allowlists{
		NetworkFence: &struct {
			AllowedDomains []string `yaml:"allowedDomains"`
		}{AllowedDomains: []string{"localhost"}},
		DependencyTyposquat: &struct {
			AllowedPackages []string `yaml:"allowedPackages"`
		}{AllowedPackages: []string{"lod-ash"}},
		ImportGuard: &struct {
			AllowedPatterns map[string][]string `yaml:"allowedPatterns"`
		}{AllowedPatterns: map[string][]string{".go": {"os/exec"}}},
	}
	out := buildAllowlistsJSON(a)
	data, _ := json.Marshal(out)
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["networkFence"]; !ok {
		t.Error("expected networkFence in output")
	}
	if dt, ok := m["dependencyTyposquat"].(map[string]interface{}); !ok {
		t.Error("expected dependencyTyposquat in output")
	} else if pkgs, ok := dt["allowedPackages"].([]interface{}); !ok || len(pkgs) != 1 {
		t.Errorf("expected allowedPackages [lod-ash], got %v", dt["allowedPackages"])
	}
	if ig, ok := m["importGuard"].(map[string]interface{}); !ok {
		t.Error("expected importGuard in output")
	} else if pats, ok := ig["allowedPatterns"].(map[string]interface{}); !ok {
		t.Errorf("expected allowedPatterns map, got %T", ig["allowedPatterns"])
	} else if goPat, ok := pats[".go"].([]interface{}); !ok || len(goPat) != 1 {
		t.Errorf("expected .go: [os/exec], got %v", pats[".go"])
	}
}

func TestHasAnyAllowlist(t *testing.T) {
	if hasAnyAllowlist(nil) {
		t.Error("nil should be false")
	}
	if hasAnyAllowlist(&config.Allowlists{}) {
		t.Error("empty allowlists should be false")
	}
	a := &config.Allowlists{
		DependencyTyposquat: &struct {
			AllowedPackages []string `yaml:"allowedPackages"`
		}{AllowedPackages: []string{"x"}},
	}
	if !hasAnyAllowlist(a) {
		t.Error("dependencyTyposquat with packages should be true")
	}
}
