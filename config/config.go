// Package config handles application configuration from file.
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	configDirName  = "clipcli"
	configFileName = "config.toml"
)

// Config holds all application settings
type Config struct {
	MaxHistory int `toml:"max_history"`
	PollMS     int `toml:"poll_ms"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxHistory: 500,
		PollMS:     300,
	}
}

// configPath returns the path to the config file
func configPath() (string, error) {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", nil
		}
		xdgConfig = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgConfig, configDirName, configFileName), nil
}

// Load loads configuration from file, returning defaults if file doesn't exist
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path, err := configPath()
	if err != nil || path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Use defaults
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), err
	}

	// Validate and apply bounds
	if cfg.MaxHistory < 10 {
		cfg.MaxHistory = 10
	}
	if cfg.MaxHistory > 10000 {
		cfg.MaxHistory = 10000
	}
	if cfg.PollMS < 50 {
		cfg.PollMS = 50
	}
	if cfg.PollMS > 5000 {
		cfg.PollMS = 5000
	}

	return cfg, nil
}

// Save writes configuration to file
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil || path == "" {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}
