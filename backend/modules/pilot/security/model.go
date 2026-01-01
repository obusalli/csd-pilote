package security

import (
	"time"

	"github.com/google/uuid"
)

// ========================================
// Firewall Rules
// ========================================

// RuleChain represents the nftables chain type
type RuleChain string

const (
	RuleChainInput       RuleChain = "INPUT"
	RuleChainOutput      RuleChain = "OUTPUT"
	RuleChainForward     RuleChain = "FORWARD"
	RuleChainPrerouting  RuleChain = "PREROUTING"
	RuleChainPostrouting RuleChain = "POSTROUTING"
)

// RuleProtocol represents the network protocol
type RuleProtocol string

const (
	RuleProtocolTCP  RuleProtocol = "TCP"
	RuleProtocolUDP  RuleProtocol = "UDP"
	RuleProtocolICMP RuleProtocol = "ICMP"
	RuleProtocolAll  RuleProtocol = "ALL"
)

// RuleAction represents the action to take
type RuleAction string

const (
	RuleActionAccept     RuleAction = "ACCEPT"
	RuleActionDrop       RuleAction = "DROP"
	RuleActionReject     RuleAction = "REJECT"
	RuleActionLog        RuleAction = "LOG"
	RuleActionMasquerade RuleAction = "MASQUERADE"
	RuleActionSnat       RuleAction = "SNAT"
	RuleActionDnat       RuleAction = "DNAT"
	RuleActionRedirect   RuleAction = "REDIRECT"
)

// ConnTrackState represents connection tracking states
type ConnTrackState string

const (
	CTStateNew         ConnTrackState = "NEW"
	CTStateEstablished ConnTrackState = "ESTABLISHED"
	CTStateRelated     ConnTrackState = "RELATED"
	CTStateInvalid     ConnTrackState = "INVALID"
)

// FirewallRule represents an individual nftables rule
type FirewallRule struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID    `json:"tenantId" gorm:"type:uuid;not null;index:idx_rule_tenant;index:idx_rule_tenant_chain_enabled"`
	Name        string       `json:"name" gorm:"not null"`
	Description string       `json:"description"`
	Chain       RuleChain    `json:"chain" gorm:"not null;default:'INPUT';index:idx_rule_tenant_chain_enabled"`
	Priority    int          `json:"priority" gorm:"default:0"`
	Protocol    RuleProtocol `json:"protocol"`
	SourceIP    string       `json:"sourceIp"`
	SourcePort  string       `json:"sourcePort"`
	DestIP      string       `json:"destIp"`
	DestPort    string       `json:"destPort"`
	Action      RuleAction   `json:"action" gorm:"not null;default:'ACCEPT'"`

	// Interface matching
	InInterface  string `json:"inInterface"`  // Input interface (iif)
	OutInterface string `json:"outInterface"` // Output interface (oif)

	// Connection tracking
	CTState string `json:"ctState"` // Connection tracking state (NEW,ESTABLISHED,RELATED,INVALID)

	// Rate limiting
	RateLimit  string `json:"rateLimit"`  // e.g., "10/second", "100/minute"
	RateBurst  int    `json:"rateBurst"`  // Burst limit
	LimitOver  string `json:"limitOver"`  // Action when limit exceeded (drop, reject)

	// NAT options (for DNAT/SNAT/REDIRECT)
	NatToAddr string `json:"natToAddr"` // Target address for DNAT/SNAT
	NatToPort string `json:"natToPort"` // Target port for DNAT/REDIRECT

	// Logging options
	LogPrefix string `json:"logPrefix"` // Prefix for log messages
	LogLevel  string `json:"logLevel"`  // Log level (emerg, alert, crit, err, warn, notice, info, debug)

	RuleExpr  string    `json:"ruleExpr"` // Raw nftables expression (advanced)
	Comment   string    `json:"comment"`
	Enabled   bool      `json:"enabled" gorm:"default:true;index:idx_rule_tenant_chain_enabled"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedBy uuid.UUID `json:"createdBy" gorm:"type:uuid"`
}

// TableName returns the table name for GORM
func (FirewallRule) TableName() string {
	return "firewall_rules"
}

// FirewallRuleInput represents input for creating/updating a firewall rule
type FirewallRuleInput struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Chain       RuleChain    `json:"chain"`
	Priority    int          `json:"priority"`
	Protocol    RuleProtocol `json:"protocol"`
	SourceIP    string       `json:"sourceIp"`
	SourcePort  string       `json:"sourcePort"`
	DestIP      string       `json:"destIp"`
	DestPort    string       `json:"destPort"`
	Action      RuleAction   `json:"action"`

	// Interface matching
	InInterface  string `json:"inInterface"`
	OutInterface string `json:"outInterface"`

	// Connection tracking
	CTState string `json:"ctState"`

	// Rate limiting
	RateLimit string `json:"rateLimit"`
	RateBurst int    `json:"rateBurst"`
	LimitOver string `json:"limitOver"`

	// NAT options
	NatToAddr string `json:"natToAddr"`
	NatToPort string `json:"natToPort"`

	// Logging options
	LogPrefix string `json:"logPrefix"`
	LogLevel  string `json:"logLevel"`

	RuleExpr string `json:"ruleExpr"`
	Comment  string `json:"comment"`
	Enabled  *bool  `json:"enabled"`
}

// FirewallRuleFilter represents filter options for listing rules
type FirewallRuleFilter struct {
	Search   *string       `json:"search"`
	Chain    *RuleChain    `json:"chain"`
	Protocol *RuleProtocol `json:"protocol"`
	Action   *RuleAction   `json:"action"`
	Enabled  *bool         `json:"enabled"`
}

// ========================================
// Firewall Profiles
// ========================================

// FirewallProfile represents a group of firewall rules
type FirewallProfile struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID      `json:"tenantId" gorm:"type:uuid;not null;index:idx_profile_tenant;index:idx_profile_tenant_default_enabled"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	IsDefault   bool           `json:"isDefault" gorm:"default:false;index:idx_profile_tenant_default_enabled"`
	Enabled     bool           `json:"enabled" gorm:"default:true;index:idx_profile_tenant_default_enabled"`

	// Default policies
	InputPolicy   string `json:"inputPolicy" gorm:"default:'drop'"`   // Default policy for input chain (accept/drop)
	OutputPolicy  string `json:"outputPolicy" gorm:"default:'accept'"` // Default policy for output chain
	ForwardPolicy string `json:"forwardPolicy" gorm:"default:'drop'"` // Default policy for forward chain

	// Features
	EnableNAT           bool `json:"enableNat" gorm:"default:false"`           // Enable NAT table
	EnableConntrack     bool `json:"enableConntrack" gorm:"default:true"`      // Enable connection tracking
	AllowLoopback       bool `json:"allowLoopback" gorm:"default:true"`        // Allow loopback traffic
	AllowEstablished    bool `json:"allowEstablished" gorm:"default:true"`     // Allow established/related connections
	AllowICMPPing       bool `json:"allowIcmpPing" gorm:"default:true"`        // Allow ICMP ping
	EnableIPv6          bool `json:"enableIpv6" gorm:"default:false"`          // Enable IPv6 support

	CreatedAt time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedBy uuid.UUID      `json:"createdBy" gorm:"type:uuid"`
	Rules     []FirewallRule `json:"rules,omitempty" gorm:"many2many:firewall_profile_rules"`
}

// TableName returns the table name for GORM
func (FirewallProfile) TableName() string {
	return "firewall_profiles"
}

// FirewallProfileRule is the join table for profile-rule relationships
type FirewallProfileRule struct {
	ProfileID uuid.UUID `json:"profileId" gorm:"type:uuid;primaryKey"`
	RuleID    uuid.UUID `json:"ruleId" gorm:"type:uuid;primaryKey"`
	SortOrder int       `json:"sortOrder" gorm:"default:0"`
}

// TableName returns the table name for GORM
func (FirewallProfileRule) TableName() string {
	return "firewall_profile_rules"
}

// FirewallProfileInput represents input for creating/updating a profile
type FirewallProfileInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsDefault   *bool    `json:"isDefault"`
	Enabled     *bool    `json:"enabled"`
	RuleIDs     []string `json:"ruleIds"` // Optional: IDs of rules to associate

	// Default policies
	InputPolicy   string `json:"inputPolicy"`
	OutputPolicy  string `json:"outputPolicy"`
	ForwardPolicy string `json:"forwardPolicy"`

	// Features
	EnableNAT        *bool `json:"enableNat"`
	EnableConntrack  *bool `json:"enableConntrack"`
	AllowLoopback    *bool `json:"allowLoopback"`
	AllowEstablished *bool `json:"allowEstablished"`
	AllowICMPPing    *bool `json:"allowIcmpPing"`
	EnableIPv6       *bool `json:"enableIpv6"`
}

// FirewallProfileFilter represents filter options for listing profiles
type FirewallProfileFilter struct {
	Search    *string `json:"search"`
	IsDefault *bool   `json:"isDefault"`
	Enabled   *bool   `json:"enabled"`
}

// ========================================
// Firewall Templates
// ========================================

// TemplateCategory represents the category of a template
type TemplateCategory string

const (
	TemplateCategoryWebServer  TemplateCategory = "WEB_SERVER"
	TemplateCategoryDatabase   TemplateCategory = "DATABASE"
	TemplateCategoryBastion    TemplateCategory = "BASTION"
	TemplateCategoryGateway    TemplateCategory = "GATEWAY"
	TemplateCategoryCustom     TemplateCategory = "CUSTOM"
)

// FirewallTemplate represents a reusable firewall template
type FirewallTemplate struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID        `json:"tenantId" gorm:"type:uuid;not null;index"`
	Name        string           `json:"name" gorm:"not null"`
	Description string           `json:"description"`
	Category    TemplateCategory `json:"category" gorm:"default:'CUSTOM'"`
	IsBuiltIn   bool             `json:"isBuiltIn" gorm:"default:false"`
	RulesJSON   string           `json:"rulesJson" gorm:"type:jsonb"` // JSON array of rule definitions
	CreatedAt   time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedBy   uuid.UUID        `json:"createdBy" gorm:"type:uuid"`
}

// TableName returns the table name for GORM
func (FirewallTemplate) TableName() string {
	return "firewall_templates"
}

// TemplateRuleDefinition represents a rule definition within a template
type TemplateRuleDefinition struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Chain       RuleChain    `json:"chain"`
	Priority    int          `json:"priority"`
	Protocol    RuleProtocol `json:"protocol"`
	SourceIP    string       `json:"sourceIp"`
	SourcePort  string       `json:"sourcePort"`
	DestIP      string       `json:"destIp"`
	DestPort    string       `json:"destPort"`
	Action      RuleAction   `json:"action"`

	// Interface matching
	InInterface  string `json:"inInterface,omitempty"`
	OutInterface string `json:"outInterface,omitempty"`

	// Connection tracking
	CTState string `json:"ctState,omitempty"`

	// Rate limiting
	RateLimit string `json:"rateLimit,omitempty"`
	RateBurst int    `json:"rateBurst,omitempty"`
	LimitOver string `json:"limitOver,omitempty"`

	// NAT options
	NatToAddr string `json:"natToAddr,omitempty"`
	NatToPort string `json:"natToPort,omitempty"`

	// Logging options
	LogPrefix string `json:"logPrefix,omitempty"`
	LogLevel  string `json:"logLevel,omitempty"`

	Comment string `json:"comment"`
}

// FirewallTemplateInput represents input for creating/updating a template
type FirewallTemplateInput struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Category    TemplateCategory         `json:"category"`
	Rules       []TemplateRuleDefinition `json:"rules"`
}

// FirewallTemplateFilter represents filter options for listing templates
type FirewallTemplateFilter struct {
	Search    *string           `json:"search"`
	Category  *TemplateCategory `json:"category"`
	IsBuiltIn *bool             `json:"isBuiltIn"`
}

// ========================================
// Firewall Deployments
// ========================================

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "PENDING"
	DeploymentStatusDeploying  DeploymentStatus = "DEPLOYING"
	DeploymentStatusApplied    DeploymentStatus = "APPLIED"
	DeploymentStatusRolledBack DeploymentStatus = "ROLLED_BACK"
	DeploymentStatusError      DeploymentStatus = "ERROR"
)

// DeploymentAction represents the type of deployment action
type DeploymentAction string

const (
	DeploymentActionApply    DeploymentAction = "APPLY"
	DeploymentActionRollback DeploymentAction = "ROLLBACK"
	DeploymentActionAudit    DeploymentAction = "AUDIT"
	DeploymentActionFlush    DeploymentAction = "FLUSH"
)

// FirewallDeployment tracks deployments of profiles to agents
type FirewallDeployment struct {
	ID            uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID         `json:"tenantId" gorm:"type:uuid;not null;index:idx_deploy_tenant;index:idx_deploy_tenant_status;index:idx_deploy_tenant_agent"`
	ProfileID     *uuid.UUID        `json:"profileId" gorm:"type:uuid"` // Optional: null for audit/flush
	AgentID       uuid.UUID         `json:"agentId" gorm:"type:uuid;not null;index:idx_deploy_tenant_agent"`
	AgentName     string            `json:"agentName"`
	Action        DeploymentAction  `json:"action" gorm:"not null;default:'APPLY'"`
	Status        DeploymentStatus  `json:"status" gorm:"default:'PENDING';index:idx_deploy_tenant_status"`
	StatusMessage string            `json:"statusMessage"`
	PlaybookID    string            `json:"playbookId"`    // csd-core playbook execution ID
	RulesSnapshot string            `json:"rulesSnapshot" gorm:"type:jsonb"` // Snapshot of rules at deploy time
	Output        string            `json:"output" gorm:"type:text"`         // Playbook output
	StartedAt     *time.Time        `json:"startedAt"`
	CompletedAt   *time.Time        `json:"completedAt"`
	CreatedAt     time.Time         `json:"createdAt" gorm:"autoCreateTime"`
	CreatedBy     uuid.UUID         `json:"createdBy" gorm:"type:uuid"`

	// Relations
	Profile *FirewallProfile `json:"profile,omitempty" gorm:"foreignKey:ProfileID"`
}

// TableName returns the table name for GORM
func (FirewallDeployment) TableName() string {
	return "firewall_deployments"
}

// DeploymentInput represents input for creating a deployment
type DeploymentInput struct {
	ProfileID string           `json:"profileId"` // Required for APPLY action
	AgentID   string           `json:"agentId"`
	Action    DeploymentAction `json:"action"`
	DryRun    bool             `json:"dryRun"` // If true, only validate without applying
}

// ProfileExport represents an exported profile with its rules
type ProfileExport struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Rules       []TemplateRuleDefinition `json:"rules"`
	ExportedAt  string                   `json:"exportedAt"`
	ExportedBy  string                   `json:"exportedBy,omitempty"`
}

// ProfileImportInput represents input for importing a profile
type ProfileImportInput struct {
	Name        string                   `json:"name,omitempty"` // Override name
	Description string                   `json:"description,omitempty"`
	Rules       []TemplateRuleDefinition `json:"rules"`
}

// FirewallDeploymentFilter represents filter options for listing deployments
type FirewallDeploymentFilter struct {
	Search    *string           `json:"search"`
	ProfileID *string           `json:"profileId"`
	AgentID   *string           `json:"agentId"`
	Action    *DeploymentAction `json:"action"`
	Status    *DeploymentStatus `json:"status"`
}
