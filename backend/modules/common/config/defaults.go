package config

// Shared default values used by multiple components (backend, CLI)
// These are the source of truth for common configuration defaults

const (
	// Database
	DefaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/csd_pilote?sslmode=prefer"

	// CSD-Core URLs
	DefaultCSDCoreURL           = "http://localhost:9090"
	DefaultCSDCoreGraphQL       = "http://localhost:9090/core/api/latest/query"
	DefaultCSDCoreServiceToken  = ""

	// Backend Server
	DefaultBackendHost = "0.0.0.0"
	DefaultBackendPort = "9092"

	// Frontend (Module Federation)
	DefaultFrontendURL       = "http://localhost:4042"
	DefaultFrontendRoutePath = "/pilote"

	// Logging
	DefaultLogLevel = "info"

	// JWT
	DefaultJWTIssuer      = "csd-pilote"
	DefaultJWTExpiryHours = 24

	// Seeds
	DefaultSeedsCoreDataPath = "data/seeds/core"
	DefaultSeedsAppDataPath  = "data/seeds/app"
)

// CommonConfigDefaults contains shared configuration keys
// These can be overridden by component-specific sections (backend.*, cli.*)
//
// Override priority (highest wins):
//  1. Component-specific: backend.database.url, cli.database.url
//  2. Common section: common.database.url
//  3. Hardcoded defaults in this file
var CommonConfigDefaults = []ConfigKeyInfo{
	// Database (shared by backend and CLI)
	{Key: "common.database.url", Type: "string", Default: DefaultDatabaseURL, Description: "PostgreSQL connection URL", Essential: true},

	// CSD-Core connection (shared by backend and CLI)
	{Key: "common.csd-core.url", Type: "string", Default: DefaultCSDCoreURL, Description: "CSD-Core backend base URL", Essential: true},
	{Key: "common.csd-core.graphql-endpoint", Type: "string", Default: DefaultCSDCoreGraphQL, Description: "CSD-Core GraphQL endpoint URL", Essential: false},
	{Key: "common.csd-core.service-token", Type: "string", Default: DefaultCSDCoreServiceToken, Description: "Service token for CSD-Core authentication", Essential: true},

	// Logging (shared by all components)
	{Key: "common.logging.level", Type: "string", Default: DefaultLogLevel, Description: "Log level: debug, info, warn, error", Essential: false},
	{Key: "common.logging.file.enabled", Type: "bool", Default: false, Description: "Enable file logging", Essential: false},
	{Key: "common.logging.file.path", Type: "string", Default: "./logs/csd-pilote.log", Description: "Log file path", Essential: false},
	{Key: "common.logging.file.max-size-mb", Type: "int", Default: 100, Description: "Max log file size in MB", Essential: false},
}
