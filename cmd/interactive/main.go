package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"hooks/internal/config"
)

// main is the program entry point. It runs the scan mode when the first command-line
// argument is "scan"; otherwise it starts the interactive mode.
func main() {
	if len(os.Args) > 1 && os.Args[1] == "scan" {
		runScan()
		return
	}
	runInteractive()
}

// runScan scans the repository and global locations for hook configuration and
// prints a concise summary to standard output.
//
// It reports the discovered config path and working directory (if any), the
// count of named hooks defined in the loaded configuration, the path to a
// generated cursor file if present, and the presence or absence of the global
// hooks file.
func runScan() {
	fmt.Println("=== Project hooks ===")
	configPath, workDir, err := config.FindConfigPath()
	if err != nil {
		fmt.Println("  (none found)")
	} else {
		fmt.Printf("  Config: %s\n", configPath)
		fmt.Printf("  Work dir: %s\n", workDir)
		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Printf("  Error loading config: %v\n", err)
		} else {
			var n int
			for _, ev := range cfg.Events() {
				for _, e := range *ev.Entries {
					if e.Name != "" {
						n++
					}
				}
			}
			fmt.Printf("  Hooks defined: %d\n", n)
		}
		cursorPath := filepath.Join(workDir, ".cursor", "hooks.json")
		if _, err := os.Stat(cursorPath); err == nil {
			fmt.Printf("  Generated: %s\n", cursorPath)
		}
	}

	fmt.Println("\n=== Global hooks ===")
	globalPath := config.GlobalHooksPath()
	if info, err := os.Stat(globalPath); err == nil && !info.IsDir() {
		fmt.Printf("  %s (present)\n", globalPath)
	} else {
		fmt.Printf("  %s (not found)\n", globalPath)
	}
}

// runInteractive starts a terminal-based interactive menu for managing project hooks.
// 
// It loads the project's configuration, presents a numbered list of defined hooks with
// their enabled status and matchers, and allows the user to toggle hooks, save changes,
// and run the repository's gen-config command. The function exits on user quit or on
// fatal configuration load errors.
func runInteractive() {
	configPath, workDir, err := config.FindConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Build flat list (event, entry index, entry) for menu
	type item struct {
		event string
		i     int
		e     *config.HookEntry
	}
	var items []item
	for _, ev := range cfg.Events() {
		for i := range *ev.Entries {
			e := &(*ev.Entries)[i]
			if e.Name != "" {
				items = append(items, item{ev.Event, i, e})
			}
		}
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- Hooks ---")
		for i, it := range items {
			status := "on "
			if it.e.Enabled != nil && !*it.e.Enabled {
				status = "off"
			}
			matcher := ""
			if it.e.Matcher != "" {
				matcher = " [" + it.e.Matcher + "]"
			}
			fmt.Printf("  %2d. [%s] %s%s (%s)\n", i+1, status, it.e.Name, matcher, it.event)
		}
		fmt.Println("\n  t <n> = toggle hook n,  s = save and run gen-config,  q = quit without saving")
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "q" || line == "quit" {
			return
		}
		if line == "s" || line == "save" {
			if err := config.Save(configPath, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "save: %v\n", err)
				continue
			}
			fmt.Println("saved", configPath)
			genConfigPath := filepath.Join(workDir, "hooks", "bin", "gen-config")
			if configPath != filepath.Join(workDir, "hooks", "config.yaml") {
				genConfigPath = filepath.Join(workDir, "bin", "gen-config")
			}
			cmd := exec.Command(genConfigPath)
			cmd.Dir = workDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "gen-config: %v\n", err)
			}
			continue
		}
		if strings.HasPrefix(line, "t ") {
			n, err := strconv.Atoi(strings.TrimSpace(line[2:]))
			if err != nil || n < 1 || n > len(items) {
				fmt.Println("invalid number")
				continue
			}
			it := &items[n-1]
			if it.e.Enabled == nil {
				f := false
				it.e.Enabled = &f
			} else {
				*it.e.Enabled = !*it.e.Enabled
			}
			status := "on"
			if it.e.Enabled != nil && !*it.e.Enabled {
				status = "off"
			}
			fmt.Printf("%s is now %s\n", it.e.Name, status)
			continue
		}
		fmt.Println("unknown command (t <n>, s, q)")
	}
}