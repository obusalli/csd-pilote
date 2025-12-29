package config

import (
	"fmt"
	"strings"

	common "csd-pilote/backend/modules/common/config"
)

// ConfigKeyInfo is re-exported from common for backward compatibility
type ConfigKeyInfo = common.ConfigKeyInfo

// ConfigDefaults contains all known configuration keys with their defaults
// This is the single source of truth for default values
// NOTE: database.url and csd-core.url are in CommonConfigDefaults (Essential: true)
var ConfigDefaults = []ConfigKeyInfo{
	// Backend Server (Essential)
	{Key: "backend.server.host", Type: "string", Default: common.DefaultBackendHost, Description: "Server bind address", Essential: true},
	{Key: "backend.server.port", Type: "string", Default: common.DefaultBackendPort, Description: "Server port", Essential: true},

	// Backend JWT (Essential)
	{Key: "backend.jwt.secret", Type: "string", Default: "", Description: "JWT signing secret (REQUIRED in production)", Essential: true},
	{Key: "backend.jwt.issuer", Type: "string", Default: common.DefaultJWTIssuer, Description: "JWT issuer name", Essential: false},
	{Key: "backend.jwt.expiry-hours", Type: "int", Default: common.DefaultJWTExpiryHours, Description: "JWT token expiry in hours", Essential: false},

	// Backend CSD-Core (Essential for service authentication)
	{Key: "backend.csd-core.url", Type: "string", Default: common.DefaultCSDCoreURL, Description: "CSD-Core backend URL override", Essential: false},
	{Key: "backend.csd-core.graphql-endpoint", Type: "string", Default: common.DefaultCSDCoreGraphQL, Description: "CSD-Core GraphQL endpoint override", Essential: false},
	{Key: "backend.csd-core.service-token", Type: "string", Default: "", Description: "Service token for CSD-Core (backend-specific)", Essential: true},

	// Backend CORS
	{Key: "backend.cors.allowed-origins", Type: "[]string", Default: []string{}, Description: "Additional CORS origins (frontend.url is auto-added)", Essential: false},
	{Key: "backend.cors.allowed-methods", Type: "[]string", Default: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, Description: "Allowed HTTP methods", Essential: false},
	{Key: "backend.cors.allowed-headers", Type: "[]string", Default: []string{"Authorization", "Content-Type"}, Description: "Allowed HTTP headers", Essential: false},

	// Backend Logging
	{Key: "backend.logging.level", Type: "string", Default: common.DefaultLogLevel, Description: "Log level override (debug, info, warn, error)", Essential: false},
	{Key: "backend.logging.file.enabled", Type: "bool", Default: false, Description: "Enable file logging", Essential: false},
	{Key: "backend.logging.file.path", Type: "string", Default: "./logs/csd-piloted.log", Description: "Log file path", Essential: false},
	{Key: "backend.logging.file.max-size-mb", Type: "int", Default: 100, Description: "Max log file size in MB", Essential: false},

	// Backend Pagination
	{Key: "backend.pagination.exact-count-threshold", Type: "int64", Default: int64(10000), Description: "Use exact COUNT(*) when estimated rows < this", Essential: false},
	{Key: "backend.pagination.estimate-count-threshold", Type: "int64", Default: int64(100000), Description: "Use pg_class estimate when estimated rows > this", Essential: false},
	{Key: "backend.pagination.always-exact-with-filters", Type: "bool", Default: true, Description: "Always use exact count when filters are applied", Essential: false},

	// Frontend (Module Federation integration)
	{Key: "frontend.url", Type: "string", Default: common.DefaultFrontendURL, Description: "Frontend URL for CORS", Essential: true},
	{Key: "frontend.remote-entry-path", Type: "string", Default: "/assets/remoteEntry.js", Description: "Module Federation remote entry path", Essential: false},
	{Key: "frontend.route-path", Type: "string", Default: common.DefaultFrontendRoutePath, Description: "Frontend route prefix", Essential: false},
}

// GetConfigKeys returns all configuration key paths
func GetConfigKeys() []string {
	keys := make([]string, len(ConfigDefaults))
	for i, info := range ConfigDefaults {
		keys[i] = info.Key
	}
	return keys
}

// GetConfigKeyInfo returns info for a specific key, or nil if not found
func GetConfigKeyInfo(key string) *ConfigKeyInfo {
	for _, info := range ConfigDefaults {
		if info.Key == key {
			return &info
		}
	}
	return nil
}

// GetEssentialKeys returns only essential configuration keys
func GetEssentialKeys() []ConfigKeyInfo {
	var essential []ConfigKeyInfo
	for _, info := range ConfigDefaults {
		if info.Essential {
			essential = append(essential, info)
		}
	}
	return essential
}

// GetDefaultInt returns the default int value for a key, or fallback if not found
func GetDefaultInt(key string, fallback int) int {
	info := GetConfigKeyInfo(key)
	if info == nil {
		return fallback
	}
	if v, ok := info.Default.(int); ok {
		return v
	}
	return fallback
}

// GetDefaultInt64 returns the default int64 value for a key, or fallback if not found
func GetDefaultInt64(key string, fallback int64) int64 {
	info := GetConfigKeyInfo(key)
	if info == nil {
		return fallback
	}
	if v, ok := info.Default.(int64); ok {
		return v
	}
	return fallback
}

// GetDefaultString returns the default string value for a key, or fallback if not found
func GetDefaultString(key string, fallback string) string {
	info := GetConfigKeyInfo(key)
	if info == nil {
		return fallback
	}
	if v, ok := info.Default.(string); ok {
		return v
	}
	return fallback
}

// GetDefaultBool returns the default bool value for a key, or fallback if not found
func GetDefaultBool(key string, fallback bool) bool {
	info := GetConfigKeyInfo(key)
	if info == nil {
		return fallback
	}
	if v, ok := info.Default.(bool); ok {
		return v
	}
	return fallback
}

// GetDefaultStringSlice returns the default []string value for a key, or fallback if not found
func GetDefaultStringSlice(key string, fallback []string) []string {
	info := GetConfigKeyInfo(key)
	if info == nil {
		return fallback
	}
	if v, ok := info.Default.([]string); ok {
		return v
	}
	return fallback
}

// GetSortedConfigDefaults returns all config defaults (already in logical order)
func GetSortedConfigDefaults() []ConfigKeyInfo {
	return ConfigDefaults
}

// GetConfigDefaultsBySection returns config defaults filtered by section prefix
func GetConfigDefaultsBySection(sectionPrefix string) []ConfigKeyInfo {
	var filtered []ConfigKeyInfo
	for _, info := range ConfigDefaults {
		if len(sectionPrefix) == 0 || strings.HasPrefix(info.Key, sectionPrefix+".") || info.Key == sectionPrefix {
			filtered = append(filtered, info)
		}
	}
	return filtered
}

// ConfigNode represents a node in the config tree for YAML generation
type ConfigNode struct {
	Children     map[string]*ConfigNode
	ChildOrder   []string // Preserve insertion order
	Value        interface{}
	DefaultValue interface{} // Default value for showing in comments
	Description  string
	IsLeaf       bool
}

// GenerateConfigYAML generates a YAML configuration file from ConfigDefaults
// The secrets map allows overriding default values (e.g., for generated secrets)
// When section is empty, includes common.* keys from CommonConfigDefaults
func GenerateConfigYAML(section string, full bool, secrets map[string]string) string {
	root := &ConfigNode{Children: make(map[string]*ConfigNode)}

	// Include common.* keys when generating full config or when section is empty/common
	if section == "" || section == "common" {
		for _, info := range common.CommonConfigDefaults {
			if !full && !info.Essential {
				continue
			}
			value := info.Default
			if secrets != nil {
				if secretVal, ok := secrets[info.Key]; ok {
					value = secretVal
				}
			}
			parts := strings.Split(info.Key, ".")
			insertIntoTree(root, parts, value, info.Default, info.Description)
		}
	}

	// Include component-specific keys
	for _, info := range ConfigDefaults {
		if section != "" && section != "common" {
			if !strings.HasPrefix(info.Key, section+".") && info.Key != section {
				continue
			}
		}
		if !full && !info.Essential {
			continue
		}

		// Use secret value if provided, otherwise use default
		value := info.Default
		if secrets != nil {
			if secretVal, ok := secrets[info.Key]; ok {
				value = secretVal
			}
		}

		parts := strings.Split(info.Key, ".")
		insertIntoTree(root, parts, value, info.Default, info.Description)
	}

	var sb strings.Builder
	sb.WriteString("# CSD Pilote Configuration\n\n")
	writeYAMLTree(&sb, root, 0)
	return sb.String()
}

func insertIntoTree(node *ConfigNode, parts []string, value interface{}, defaultValue interface{}, description string) {
	if len(parts) == 0 {
		return
	}
	key := parts[0]
	child, exists := node.Children[key]
	if !exists {
		child = &ConfigNode{Children: make(map[string]*ConfigNode)}
		node.Children[key] = child
		node.ChildOrder = append(node.ChildOrder, key)
	}
	if len(parts) == 1 {
		child.Value = value
		child.DefaultValue = defaultValue
		child.Description = description
		child.IsLeaf = true
	} else {
		insertIntoTree(child, parts[1:], value, defaultValue, description)
	}
}

func writeYAMLTree(sb *strings.Builder, node *ConfigNode, indent int) {
	indentStr := strings.Repeat("  ", indent)

	for _, key := range node.ChildOrder {
		child := node.Children[key]
		if child.IsLeaf {
			sb.WriteString(indentStr)
			sb.WriteString(key)
			sb.WriteString(":")

			comment := buildComment(child.Description, child.DefaultValue)

			isNonEmptyArray := false
			switch arr := child.Value.(type) {
			case []string:
				isNonEmptyArray = len(arr) > 0
			case []interface{}:
				isNonEmptyArray = len(arr) > 0
			}

			if isNonEmptyArray {
				sb.WriteString("  # ")
				sb.WriteString(comment)
				sb.WriteString("\n")
				writeYAMLValue(sb, child.Value, indent)
			} else {
				sb.WriteString(" ")
				writeYAMLValue(sb, child.Value, indent)
				sb.WriteString("  # ")
				sb.WriteString(comment)
				sb.WriteString("\n")
			}
		} else {
			sb.WriteString(indentStr)
			sb.WriteString(key)
			sb.WriteString(":\n")
			writeYAMLTree(sb, child, indent+1)
		}
	}
}

func buildComment(description string, defaultValue interface{}) string {
	defaultStr := formatValueForComment(defaultValue)
	if description != "" {
		return fmt.Sprintf("%s (default: %s)", description, defaultStr)
	}
	return fmt.Sprintf("default: %s", defaultStr)
}

func formatValueForComment(value interface{}) string {
	switch v := value.(type) {
	case string:
		if v == "" {
			return `""`
		}
		return v
	case []string:
		if len(v) == 0 {
			return "[]"
		}
		return fmt.Sprintf("[%s]", strings.Join(v, ", "))
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", value)
	}
}

func writeYAMLValue(sb *strings.Builder, value interface{}, indent int) {
	switch v := value.(type) {
	case string:
		if v == "" {
			sb.WriteString(`""`)
		} else if needsQuoting(v) {
			sb.WriteString(`"`)
			sb.WriteString(strings.ReplaceAll(v, `"`, `\"`))
			sb.WriteString(`"`)
		} else {
			sb.WriteString(v)
		}
	case int:
		sb.WriteString(fmt.Sprintf("%d", v))
	case int64:
		sb.WriteString(fmt.Sprintf("%d", v))
	case bool:
		if v {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case []string:
		if len(v) == 0 {
			sb.WriteString("[]")
		} else {
			indentStr := strings.Repeat("  ", indent+1)
			for _, item := range v {
				sb.WriteString(indentStr)
				sb.WriteString("- ")
				if needsQuoting(item) {
					sb.WriteString(`"`)
					sb.WriteString(strings.ReplaceAll(item, `"`, `\"`))
					sb.WriteString(`"`)
				} else {
					sb.WriteString(item)
				}
				sb.WriteString("\n")
			}
		}
	default:
		sb.WriteString(fmt.Sprintf("%v", value))
	}
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	special := []string{":", "#", "[", "]", "{", "}", ",", "&", "*", "!", "|", ">", "'", `"`, "%", "@", "`", " ", "\t", "\n"}
	for _, char := range special {
		if strings.Contains(s, char) {
			return true
		}
	}
	lower := strings.ToLower(s)
	keywords := []string{"true", "false", "yes", "no", "on", "off", "null", "~"}
	for _, kw := range keywords {
		if lower == kw {
			return true
		}
	}
	return false
}
