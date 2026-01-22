package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Sharing SharingConfig `yaml:"sharing"`
	Sync    SyncConfig    `yaml:"sync"`
	Logging LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

type SharingConfig struct {
	Level           string   `yaml:"level"` // none, metadata, full
	ExcludeProjects []string `yaml:"exclude_projects"`
	AnonymizePaths  bool     `yaml:"anonymize_paths"`
}

type SyncConfig struct {
	Interval      int `yaml:"interval"` // seconds
	RetryAttempts int `yaml:"retry_attempts"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			URL: "https://insights.dkd.internal",
		},
		Sharing: SharingConfig{
			Level:          "metadata",
			AnonymizePaths: true,
		},
		Sync: SyncConfig{
			Interval:      300,
			RetryAttempts: 3,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

// ConfigPath returns the default config file path
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude-insights", "config.yaml")
}

// StatePath returns the default state file path
func StatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "claude-insights", "synced.json")
}

// ClaudeLogsPath returns the Claude Code logs directory
func ClaudeLogsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

// Load reads config from the given path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes config to the given path
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Validate checks if config is valid
func (c *Config) Validate() error {
	if c.Server.URL == "" {
		return ErrMissingServerURL
	}
	if c.Server.APIKey == "" {
		return ErrMissingAPIKey
	}
	if c.Sharing.Level != "none" && c.Sharing.Level != "metadata" && c.Sharing.Level != "full" {
		return ErrInvalidShareLevel
	}
	return nil
}

// Errors
var (
	ErrMissingServerURL  = &ConfigError{"server.url is required"}
	ErrMissingAPIKey     = &ConfigError{"server.api_key is required"}
	ErrInvalidShareLevel = &ConfigError{"sharing.level must be none, metadata, or full"}
)

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
