package config

// ConfigKeyInfo describes a configuration key with its metadata
// Used by CLI and backend for config generation and validation
type ConfigKeyInfo struct {
	Key         string      // Full key path (e.g., "backend.server.port")
	Type        string      // Type: string, int, bool, []string
	Default     interface{} // Default value
	Description string      // Description for help
	Essential   bool        // True if key is essential for basic operation
}
