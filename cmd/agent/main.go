package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dkd/claude-insights-agent/internal/config"
	"github.com/dkd/claude-insights-agent/internal/watcher"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		cmdInit()
	case "run":
		cmdRun()
	case "sync":
		cmdSync()
	case "status":
		cmdStatus()
	case "version", "-v", "--version":
		fmt.Printf("claude-insights-agent v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("claude-insights-agent - Sync Claude Code sessions to team server")
	fmt.Println()
	fmt.Println("Usage: claude-insights-agent <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init      Initialize configuration (interactive)")
	fmt.Println("  run       Start continuous sync daemon")
	fmt.Println("  sync      Run one-time sync")
	fmt.Println("  status    Show sync status")
	fmt.Println("  version   Show version")
	fmt.Println("  help      Show this help")
}

func cmdInit() {
	cfgPath := config.ConfigPath()

	// Check if config already exists
	if _, err := os.Stat(cfgPath); err == nil {
		fmt.Printf("Config already exists at %s\n", cfgPath)
		fmt.Print("Overwrite? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("Aborted")
			return
		}
	}

	cfg := config.DefaultConfig()

	reader := bufio.NewReader(os.Stdin)

	// Server URL
	fmt.Printf("Server URL [%s]: ", cfg.Server.URL)
	if url, _ := reader.ReadString('\n'); strings.TrimSpace(url) != "" {
		cfg.Server.URL = strings.TrimSpace(url)
	}

	// API Key
	fmt.Print("API Key: ")
	apiKey, _ := reader.ReadString('\n')
	cfg.Server.APIKey = strings.TrimSpace(apiKey)

	// Share level
	fmt.Println()
	fmt.Println("Share level options:")
	fmt.Println("  none     - Don't share anything (agent paused)")
	fmt.Println("  metadata - Share session stats, tokens, tools (no content)")
	fmt.Println("  full     - Share everything including message content")
	fmt.Printf("Share level [%s]: ", cfg.Sharing.Level)
	if level, _ := reader.ReadString('\n'); strings.TrimSpace(level) != "" {
		cfg.Sharing.Level = strings.TrimSpace(level)
	}

	// Anonymize paths
	fmt.Printf("Anonymize project paths? [Y/n]: ")
	if anon, _ := reader.ReadString('\n'); strings.TrimSpace(strings.ToLower(anon)) == "n" {
		cfg.Sharing.AnonymizePaths = false
	}

	// Sync interval
	fmt.Printf("Sync interval in seconds [%d]: ", cfg.Sync.Interval)
	if interval, _ := reader.ReadString('\n'); strings.TrimSpace(interval) != "" {
		var i int
		if _, err := fmt.Sscanf(strings.TrimSpace(interval), "%d", &i); err == nil && i > 0 {
			cfg.Sync.Interval = i
		}
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Save
	if err := cfg.Save(cfgPath); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nConfiguration saved to %s\n", cfgPath)
	fmt.Println("Run 'claude-insights-agent run' to start syncing")
}

func cmdRun() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Run 'claude-insights-agent init' to create config")
		os.Exit(1)
	}

	logger := setupLogger(cfg)

	w := watcher.New(cfg, logger)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Println("Shutting down...")
		w.Stop()
	}()

	fmt.Println("Starting claude-insights-agent...")
	fmt.Printf("Syncing to %s (share level: %s)\n", cfg.Server.URL, cfg.Sharing.Level)
	fmt.Println("Press Ctrl+C to stop")

	if err := w.Start(); err != nil {
		logger.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func cmdSync() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Run 'claude-insights-agent init' to create config")
		os.Exit(1)
	}

	logger := setupLogger(cfg)

	w := watcher.New(cfg, logger)

	fmt.Println("Running one-time sync...")
	if err := w.SyncOnce(); err != nil {
		fmt.Printf("Sync error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Sync complete")
}

func cmdStatus() {
	cfgPath := config.ConfigPath()
	statePath := config.StatePath()

	// Check config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Println("Status: NOT CONFIGURED")
		fmt.Printf("Config file: %s (missing)\n", cfgPath)
		fmt.Println("Run 'claude-insights-agent init' to create config")
		return
	}

	fmt.Println("Status: CONFIGURED")
	fmt.Printf("Config file: %s\n", cfgPath)
	fmt.Printf("Server: %s\n", cfg.Server.URL)
	fmt.Printf("Share level: %s\n", cfg.Sharing.Level)
	fmt.Printf("Sync interval: %ds\n", cfg.Sync.Interval)
	fmt.Printf("Anonymize paths: %v\n", cfg.Sharing.AnonymizePaths)
	fmt.Println()

	// Check state
	logger := log.New(os.Stderr, "", 0)
	w := watcher.New(cfg, logger)
	stats := w.GetStats()

	fmt.Printf("State file: %s\n", statePath)
	fmt.Printf("Sessions synced: %d\n", stats.TotalSynced)
	if !stats.LastSync.IsZero() {
		fmt.Printf("Last sync: %s\n", stats.LastSync.Format("2006-01-02 15:04:05"))
	}
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func setupLogger(cfg *config.Config) *log.Logger {
	var output *os.File

	if cfg.Logging.File != "" {
		// Expand ~ to home directory
		logPath := cfg.Logging.File
		if strings.HasPrefix(logPath, "~") {
			home, _ := os.UserHomeDir()
			logPath = strings.Replace(logPath, "~", home, 1)
		}

		// Ensure directory exists
		os.MkdirAll(strings.TrimSuffix(logPath, "/"+strings.Split(logPath, "/")[len(strings.Split(logPath, "/"))-1]), 0755)

		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Warning: could not open log file: %v\n", err)
			output = os.Stdout
		} else {
			output = f
		}
	} else {
		output = os.Stdout
	}

	return log.New(output, "[insights] ", log.Ldate|log.Ltime)
}
