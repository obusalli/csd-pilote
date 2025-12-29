package config

import (
	"fmt"
	"strings"

	common "csd-pilote/backend/modules/common/config"
)

// Re-export ConfigKeyInfo from common for convenience
type ConfigKeyInfo = common.ConfigKeyInfo

// ConfigDefaults contains all known CLI configuration keys with their defaults
// NOTE: database.url and csd-core.url are in CommonConfigDefaults (Essential: true)
var ConfigDefaults = []ConfigKeyInfo{
	// CLI Core
	{Key: "cli.dev-mode", Type: "bool", Default: false, Description: "Enable development mode", Essential: false},

	// CLI CSD-Core connection
	{Key: "cli.csd-core.url", Type: "string", Default: common.DefaultCSDCoreURL, Description: "CSD-Core backend URL override", Essential: false},
	{Key: "cli.csd-core.graphql-endpoint", Type: "string", Default: common.DefaultCSDCoreGraphQL, Description: "CSD-Core GraphQL endpoint override", Essential: true},
	{Key: "cli.csd-core.service-token", Type: "string", Default: "", Description: "Service token for CSD-Core (CLI-specific)", Essential: false},

	// CLI Seeds
	{Key: "cli.seeds.core-path", Type: "string", Default: common.DefaultSeedsCoreDataPath, Description: "Path to core seed data", Essential: false},
	{Key: "cli.seeds.app-path", Type: "string", Default: common.DefaultSeedsAppDataPath, Description: "Path to app seed data", Essential: false},

	// CLI Logging
	{Key: "cli.logging.level", Type: "string", Default: common.DefaultLogLevel, Description: "CLI log level override", Essential: false},
	{Key: "cli.logging.file.enabled", Type: "bool", Default: false, Description: "Enable file logging", Essential: false},
	{Key: "cli.logging.file.path", Type: "string", Default: "./logs/csd-pilotectl.log", Description: "Log file path", Essential: false},
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

// GetDefaultInt returns the default int value for a key
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

// GetDefaultString returns the default string value for a key
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

// GetDefaultBool returns the default bool value for a key
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

// GetDefaultStringSlice returns the default []string value for a key
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
	Children   map[string]*ConfigNode
	ChildOrder []string
	Value      interface{}
	Default    interface{}
	Desc       string
	IsLeaf     bool
}

// GenerateConfigYAML generates a YAML configuration from defaults
// If essentialOnly is true, only essential keys are included
// Includes common.* keys from CommonConfigDefaults
func GenerateConfigYAML(essentialOnly bool) string {
	root := &ConfigNode{Children: make(map[string]*ConfigNode)}

	// Include common.* keys first
	for _, info := range common.CommonConfigDefaults {
		if essentialOnly && !info.Essential {
			continue
		}
		parts := strings.Split(info.Key, ".")
		insertIntoTree(root, parts, info.Default, info.Default, info.Description)
	}

	// Include CLI-specific keys
	for _, info := range ConfigDefaults {
		if essentialOnly && !info.Essential {
			continue
		}
		parts := strings.Split(info.Key, ".")
		insertIntoTree(root, parts, info.Default, info.Default, info.Description)
	}

	var sb strings.Builder
	sb.WriteString("# CSD Pilote CLI Configuration\n\n")
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
		child.Default = defaultValue
		child.Desc = description
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
			// Write comment
			if child.Desc != "" {
				sb.WriteString(indentStr)
				sb.WriteString("# ")
				sb.WriteString(child.Desc)
				sb.WriteString("\n")
			}
			sb.WriteString(indentStr)
			sb.WriteString(key)
			sb.WriteString(": ")
			writeYAMLValue(sb, child.Value, indent)
			sb.WriteString("\n")
		} else {
			sb.WriteString(indentStr)
			sb.WriteString(key)
			sb.WriteString(":\n")
			writeYAMLTree(sb, child, indent+1)
		}
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
			sb.WriteString("\n")
			itemIndent := strings.Repeat("  ", indent+1)
			for _, item := range v {
				sb.WriteString(itemIndent)
				sb.WriteString("- ")
				if needsQuoting(item) {
					sb.WriteString(`"`)
					sb.WriteString(item)
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
	special := []string{":", "#", "[", "]", "{", "}", ",", "&", "*", "!", "|", ">", "'", `"`, "%", "@", "`", " "}
	for _, char := range special {
		if strings.Contains(s, char) {
			return true
		}
	}
	lower := strings.ToLower(s)
	keywords := []string{"true", "false", "yes", "no", "on", "off", "null"}
	for _, kw := range keywords {
		if lower == kw {
			return true
		}
	}
	return false
}
