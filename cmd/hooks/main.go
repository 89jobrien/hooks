package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed config_default.yaml
var defaultConfigYAML []byte

func main() {
	if len(os.Args) < 2 || os.Args[1] != "init" {
		fmt.Fprintf(os.Stderr, "usage: hooks init [path]\n")
		fmt.Fprintf(os.Stderr, "  Initialize a repo with .hooks/config.yaml. Path defaults to current directory.\n")
		os.Exit(1)
	}

	target := "."
	if len(os.Args) > 2 {
		target = os.Args[2]
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}

	hooksDir := filepath.Join(absTarget, ".hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(hooksDir, "config.yaml")
	if err := os.WriteFile(configPath, defaultConfigYAML, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("wrote", configPath)

	// Run gen-config from target so .cursor/ and .claude/ are generated (skip binary check).
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
	binDir := filepath.Dir(exe)
	genConfigName := "gen-config"
	if runtime.GOOS == "windows" {
		genConfigName = "gen-config.exe"
	}
	genConfig := filepath.Join(binDir, genConfigName)
	if info, err := os.Stat(genConfig); err != nil || info.IsDir() {
		fmt.Fprintf(os.Stderr, "init: gen-config not found next to hooks binary; run gen-config -skip-validate from repo root after installing binaries\n")
		return
	}
	cmd := exec.Command(genConfig, "-skip-validate")
	cmd.Dir = absTarget
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "init: gen-config: %v\n", err)
		os.Exit(1)
	}
}
