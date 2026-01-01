package config

import (
	"fmt"
	"os"
	"path/filepath"

	common "csd-pilote/backend/modules/common/config"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration (final merged config)
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	CSDCore    CSDCoreConfig    `yaml:"csd-core"`
	Frontend   FrontendConfig   `yaml:"frontend"`
	JWT        JWTConfig        `yaml:"jwt"`
	CORS       CORSConfig       `yaml:"cors"`
	Logging    LoggingConfig    `yaml:"logging"`
	CLI        CLIConfig        `yaml:"cli"`
	Pagination PaginationConfig `yaml:"pagination"`
	Limits     LimitsConfig     `yaml:"limits"`
}

// PaginationConfig configures pagination and count strategies
type PaginationConfig struct {
	DefaultLimit           int   `yaml:"default_limit"`
	MaxLimit               int   `yaml:"max_limit"`
	ExactCountThreshold    int64 `yaml:"exact_count_threshold"`
	EstimateCountThreshold int64 `yaml:"estimate_count_threshold"`
	AlwaysExactWithFilters bool  `yaml:"always_exact_with_filters"`
}

// LimitsConfig configures various resource limits
type LimitsConfig struct {
	MaxNodesPerCluster          int `yaml:"max_nodes_per_cluster"`
	ClusterDeploymentTimeout    int `yaml:"cluster_deployment_timeout_minutes"`
	HypervisorDeploymentTimeout int `yaml:"hypervisor_deployment_timeout_minutes"`
	FirewallDeploymentTimeout   int `yaml:"firewall_deployment_timeout_minutes"`
}

// RawConfig represents the YAML file structure with common/backend/frontend/cli sections
type RawConfig struct {
	Common   CommonConfig   `yaml:"common"`
	Backend  BackendConfig  `yaml:"backend"`
	Frontend FrontendConfig `yaml:"frontend"`
	CLI      CLIConfig      `yaml:"cli"`
}

type CommonConfig struct {
	Database DatabaseConfig `yaml:"database"`
	CSDCore  CSDCoreConfig  `yaml:"csd-core"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type BackendConfig struct {
	Server  ServerConfig  `yaml:"server"`
	CSDCore CSDCoreConfig `yaml:"csd-core"`
	JWT     JWTConfig     `yaml:"jwt"`
	CORS    CORSConfig    `yaml:"cors"`
	Logging LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	URL string `yaml:"url"`
}

type CSDCoreConfig struct {
	URL             string `yaml:"url"`
	GraphQLEndpoint string `yaml:"graphql-endpoint"`
	ServiceToken    string `yaml:"service-token"`
}

// FrontendConfig holds frontend integration settings for Module Federation
type FrontendConfig struct {
	URL             string `yaml:"url"`               // e.g., http://localhost:4042
	RemoteEntryPath string `yaml:"remote-entry-path"` // e.g., /assets/remoteEntry.js
	RoutePath       string `yaml:"route-path"`        // e.g., /pilote
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	Issuer      string `yaml:"issuer"`
	ExpiryHours int    `yaml:"expiry-hours"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed-origins"`
	AllowedMethods []string `yaml:"allowed-methods"`
	AllowedHeaders []string `yaml:"allowed-headers"`
}

type LoggingConfig struct {
	Level string            `yaml:"level"`
	File  LoggingFileConfig `yaml:"file"`
}

type LoggingFileConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Path      string `yaml:"path"`
	MaxSizeMB int    `yaml:"max-size-mb"`
}

type CLIConfig struct {
	Seeds SeedsConfig `yaml:"seeds"`
}

type SeedsConfig struct {
	CorePath string `yaml:"core-path"`
	AppPath  string `yaml:"app-path"`
}

var globalConfig *Config

// Load loads configuration from a YAML file
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		// Try default locations
		candidates := []string{
			"csd-pilote.yaml",
			"backend/csd-pilote.yaml",
			filepath.Join(os.Getenv("HOME"), ".config", "csd-pilote", "csd-pilote.yaml"),
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
				break
			}
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("no configuration file found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var rawCfg RawConfig
	if err := yaml.Unmarshal(data, &rawCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge common with backend-specific config
	cfg := mergeConfig(rawCfg)

	// Set defaults from ConfigDefaults (single source of truth)
	if cfg.Server.Host == "" {
		cfg.Server.Host = common.DefaultBackendHost
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = common.DefaultBackendPort
	}
	if cfg.CSDCore.URL == "" {
		cfg.CSDCore.URL = common.DefaultCSDCoreURL
	}
	if cfg.CSDCore.GraphQLEndpoint == "" {
		cfg.CSDCore.GraphQLEndpoint = common.DefaultCSDCoreGraphQL
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = common.DefaultJWTIssuer
	}
	if cfg.JWT.ExpiryHours == 0 {
		cfg.JWT.ExpiryHours = common.DefaultJWTExpiryHours
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = common.DefaultLogLevel
	}
	if cfg.Frontend.URL == "" {
		cfg.Frontend.URL = common.DefaultFrontendURL
	}
	if cfg.Frontend.RoutePath == "" {
		cfg.Frontend.RoutePath = common.DefaultFrontendRoutePath
	}

	// Pagination defaults
	if cfg.Pagination.DefaultLimit == 0 {
		cfg.Pagination.DefaultLimit = 20
	}
	if cfg.Pagination.MaxLimit == 0 {
		cfg.Pagination.MaxLimit = 100
	}
	if cfg.Pagination.ExactCountThreshold == 0 {
		cfg.Pagination.ExactCountThreshold = GetDefaultInt64("backend.pagination.exact-count-threshold", 10000)
	}
	if cfg.Pagination.EstimateCountThreshold == 0 {
		cfg.Pagination.EstimateCountThreshold = GetDefaultInt64("backend.pagination.estimate-count-threshold", 100000)
	}

	// Limits defaults
	if cfg.Limits.MaxNodesPerCluster == 0 {
		cfg.Limits.MaxNodesPerCluster = 1000
	}
	if cfg.Limits.ClusterDeploymentTimeout == 0 {
		cfg.Limits.ClusterDeploymentTimeout = 30 // minutes
	}
	if cfg.Limits.HypervisorDeploymentTimeout == 0 {
		cfg.Limits.HypervisorDeploymentTimeout = 15 // minutes
	}
	if cfg.Limits.FirewallDeploymentTimeout == 0 {
		cfg.Limits.FirewallDeploymentTimeout = 5 // minutes
	}

	globalConfig = &cfg
	return &cfg, nil
}

// mergeConfig merges common config with backend-specific overrides
func mergeConfig(raw RawConfig) Config {
	cfg := Config{
		// Start with common values
		Database: raw.Common.Database,
		Logging:  raw.Common.Logging,
		CSDCore:  raw.Common.CSDCore,

		// Backend-specific values
		Server:   raw.Backend.Server,
		JWT:      raw.Backend.JWT,
		CORS:     raw.Backend.CORS,
		Frontend: raw.Frontend,
		CLI:      raw.CLI,
	}

	// Override common with backend-specific if set
	if raw.Backend.CSDCore.ServiceToken != "" {
		cfg.CSDCore.ServiceToken = raw.Backend.CSDCore.ServiceToken
	}
	if raw.Backend.CSDCore.URL != "" {
		cfg.CSDCore.URL = raw.Backend.CSDCore.URL
	}
	if raw.Backend.CSDCore.GraphQLEndpoint != "" {
		cfg.CSDCore.GraphQLEndpoint = raw.Backend.CSDCore.GraphQLEndpoint
	}

	// Override logging if backend-specific is set
	if raw.Backend.Logging.Level != "" {
		cfg.Logging.Level = raw.Backend.Logging.Level
	}
	if raw.Backend.Logging.File.Path != "" {
		cfg.Logging.File = raw.Backend.Logging.File
	}

	return cfg
}

// GetConfig returns the global configuration
func GetConfig() *Config {
	return globalConfig
}

// SetConfig sets the global configuration
func SetConfig(cfg *Config) {
	globalConfig = cfg
}
