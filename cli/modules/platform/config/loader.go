package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration (final merged config)
type Config struct {
	DevMode  bool           `yaml:"dev-mode"`
	Database DatabaseConfig `yaml:"database"`
	CSDCore  CSDCoreConfig  `yaml:"csd-core"`
	Logging  LoggingConfig  `yaml:"logging"`
	Seeds    SeedsConfig    `yaml:"seeds"`
}

// RawConfig represents the YAML file structure with common/cli sections
type RawConfig struct {
	Common CommonConfig `yaml:"common"`
	CLI    CLIConfig    `yaml:"cli"`
}

type CommonConfig struct {
	Database DatabaseConfig `yaml:"database"`
	CSDCore  CSDCoreConfig  `yaml:"csd-core"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type CLIConfig struct {
	DevMode  bool          `yaml:"dev-mode"`
	CSDCore  CSDCoreConfig `yaml:"csd-core"`
	Seeds    SeedsConfig   `yaml:"seeds"`
	Logging  LoggingConfig `yaml:"logging"`
}

type DatabaseConfig struct {
	URL string `yaml:"url"`
}

type CSDCoreConfig struct {
	URL             string `yaml:"url"`
	GraphQLEndpoint string `yaml:"graphql-endpoint"`
	ServiceToken    string `yaml:"service-token"`
}

type LoggingConfig struct {
	Level string            `yaml:"level"`
	File  LoggingFileConfig `yaml:"file"`
}

type LoggingFileConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type SeedsConfig struct {
	CorePath string `yaml:"core-path"`
	AppPath  string `yaml:"app-path"`
}

var globalConfig *Config
var configPath string

// LoadConfig loads configuration from file
func LoadConfig() (*Config, error) {
	path := findConfigPath()
	if path == "" {
		return nil, fmt.Errorf("no configuration file found")
	}

	configPath = path
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var rawCfg RawConfig
	if err := yaml.Unmarshal(data, &rawCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge common with cli-specific config
	cfg := mergeConfig(rawCfg)

	// Apply defaults
	applyDefaults(&cfg)

	globalConfig = &cfg
	return &cfg, nil
}

// mergeConfig merges common config with cli-specific overrides
func mergeConfig(raw RawConfig) Config {
	cfg := Config{
		// Start with common values
		Database: raw.Common.Database,
		Logging:  raw.Common.Logging,
		CSDCore:  raw.Common.CSDCore,

		// CLI-specific values
		DevMode: raw.CLI.DevMode,
		Seeds:   raw.CLI.Seeds,
	}

	// Override common with cli-specific if set
	if raw.CLI.CSDCore.ServiceToken != "" {
		cfg.CSDCore.ServiceToken = raw.CLI.CSDCore.ServiceToken
	}
	if raw.CLI.CSDCore.URL != "" {
		cfg.CSDCore.URL = raw.CLI.CSDCore.URL
	}
	if raw.CLI.CSDCore.GraphQLEndpoint != "" {
		cfg.CSDCore.GraphQLEndpoint = raw.CLI.CSDCore.GraphQLEndpoint
	}

	// Override logging if cli-specific is set
	if raw.CLI.Logging.Level != "" {
		cfg.Logging.Level = raw.CLI.Logging.Level
	}
	if raw.CLI.Logging.File.Path != "" {
		cfg.Logging.File = raw.CLI.Logging.File
	}

	return cfg
}

// findConfigPath finds the configuration file
func findConfigPath() string {
	// Check command line flag first
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}

	// Default locations
	candidates := []string{
		"csd-pilote.yaml",
		"../backend/csd-pilote.yaml",
		filepath.Join(os.Getenv("HOME"), ".config", "csd-pilote", "csd-pilote.yaml"),
		"/etc/csd-pilote/csd-pilote.yaml",
	}

	// Also check executable directory
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append([]string{filepath.Join(exeDir, "csd-pilote.yaml")}, candidates...)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// applyDefaults sets default values for missing configuration
func applyDefaults(cfg *Config) {
	if cfg.Seeds.CorePath == "" {
		cfg.Seeds.CorePath = "data/seeds/core"
	}
	if cfg.Seeds.AppPath == "" {
		cfg.Seeds.AppPath = "data/seeds/app"
	}
	if cfg.CSDCore.URL == "" {
		cfg.CSDCore.URL = "http://localhost:9090"
	}
	if cfg.CSDCore.GraphQLEndpoint == "" {
		cfg.CSDCore.GraphQLEndpoint = "/core/api/latest/query"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
}

// GetConfig returns the global configuration
func GetConfig() *Config {
	return globalConfig
}

// SetConfig sets the global configuration
func SetConfig(cfg *Config) {
	globalConfig = cfg
}

// GetConfigPath returns the path of the loaded config file
func GetConfigPath() string {
	return configPath
}
