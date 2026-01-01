package security

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/platform/config"
	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/events"
	"csd-pilote/backend/modules/platform/pagination"
)

// Service handles business logic for firewall security
type Service struct {
	repo   *Repository
	client *csdcore.Client
}

// NewService creates a new security service
func NewService() *Service {
	return &Service{
		repo:   NewRepository(),
		client: csdcore.GetClient(),
	}
}

// ========================================
// Firewall Rules
// ========================================

// CreateRule creates a new firewall rule
func (s *Service) CreateRule(ctx context.Context, token string, tenantID, userID uuid.UUID, input *FirewallRuleInput) (*FirewallRule, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	rule := &FirewallRule{
		TenantID:     tenantID,
		Name:         input.Name,
		Description:  input.Description,
		Chain:        input.Chain,
		Priority:     input.Priority,
		Protocol:     input.Protocol,
		SourceIP:     input.SourceIP,
		SourcePort:   input.SourcePort,
		DestIP:       input.DestIP,
		DestPort:     input.DestPort,
		Action:       input.Action,
		InInterface:  input.InInterface,
		OutInterface: input.OutInterface,
		CTState:      input.CTState,
		RateLimit:    input.RateLimit,
		RateBurst:    input.RateBurst,
		LimitOver:    input.LimitOver,
		NatToAddr:    input.NatToAddr,
		NatToPort:    input.NatToPort,
		LogPrefix:    input.LogPrefix,
		LogLevel:     input.LogLevel,
		RuleExpr:     input.RuleExpr,
		Comment:      input.Comment,
		Enabled:      enabled,
		CreatedBy:    userID,
	}

	// Set defaults
	if rule.Chain == "" {
		rule.Chain = RuleChainInput
	}
	if rule.Action == "" {
		rule.Action = RuleActionAccept
	}

	if err := s.repo.CreateRule(rule); err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallRuleCreated,
		tenantID,
		rule.ID.String(),
		map[string]interface{}{
			"name":   rule.Name,
			"chain":  rule.Chain,
			"action": rule.Action,
		},
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.rule.created",
		ResourceType: "firewall_rule",
		ResourceID:   rule.ID.String(),
		Details: map[string]interface{}{
			"name":     rule.Name,
			"chain":    rule.Chain,
			"action":   rule.Action,
			"protocol": rule.Protocol,
			"enabled":  rule.Enabled,
		},
	})

	return rule, nil
}

// GetRule retrieves a rule by ID
func (s *Service) GetRule(ctx context.Context, tenantID, id uuid.UUID) (*FirewallRule, error) {
	return s.repo.GetRuleByID(tenantID, id)
}

// ListRules retrieves all rules for a tenant
func (s *Service) ListRules(ctx context.Context, tenantID uuid.UUID, filter *FirewallRuleFilter, limit, offset int) ([]FirewallRule, int64, error) {
	p := pagination.Normalize(limit, offset)
	return s.repo.ListRules(tenantID, filter, p.Limit, p.Offset)
}

// UpdateRule updates a firewall rule
func (s *Service) UpdateRule(ctx context.Context, token string, tenantID, id uuid.UUID, input *FirewallRuleInput) (*FirewallRule, error) {
	rule, err := s.repo.GetRuleByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		rule.Name = input.Name
	}
	if input.Description != "" {
		rule.Description = input.Description
	}
	if input.Chain != "" {
		rule.Chain = input.Chain
	}
	if input.Priority != 0 {
		rule.Priority = input.Priority
	}
	if input.Protocol != "" {
		rule.Protocol = input.Protocol
	}
	if input.SourceIP != "" {
		rule.SourceIP = input.SourceIP
	}
	if input.SourcePort != "" {
		rule.SourcePort = input.SourcePort
	}
	if input.DestIP != "" {
		rule.DestIP = input.DestIP
	}
	if input.DestPort != "" {
		rule.DestPort = input.DestPort
	}
	if input.Action != "" {
		rule.Action = input.Action
	}
	// Interface matching
	if input.InInterface != "" {
		rule.InInterface = input.InInterface
	}
	if input.OutInterface != "" {
		rule.OutInterface = input.OutInterface
	}
	// Connection tracking
	if input.CTState != "" {
		rule.CTState = input.CTState
	}
	// Rate limiting
	if input.RateLimit != "" {
		rule.RateLimit = input.RateLimit
	}
	if input.RateBurst != 0 {
		rule.RateBurst = input.RateBurst
	}
	if input.LimitOver != "" {
		rule.LimitOver = input.LimitOver
	}
	// NAT options
	if input.NatToAddr != "" {
		rule.NatToAddr = input.NatToAddr
	}
	if input.NatToPort != "" {
		rule.NatToPort = input.NatToPort
	}
	// Logging options
	if input.LogPrefix != "" {
		rule.LogPrefix = input.LogPrefix
	}
	if input.LogLevel != "" {
		rule.LogLevel = input.LogLevel
	}
	if input.RuleExpr != "" {
		rule.RuleExpr = input.RuleExpr
	}
	if input.Comment != "" {
		rule.Comment = input.Comment
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}

	if err := s.repo.UpdateRule(rule); err != nil {
		return nil, fmt.Errorf("failed to update rule: %w", err)
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallRuleUpdated,
		tenantID,
		rule.ID.String(),
		map[string]interface{}{
			"name":   rule.Name,
			"chain":  rule.Chain,
			"action": rule.Action,
		},
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.rule.updated",
		ResourceType: "firewall_rule",
		ResourceID:   rule.ID.String(),
		Details: map[string]interface{}{
			"name":     rule.Name,
			"chain":    rule.Chain,
			"action":   rule.Action,
			"protocol": rule.Protocol,
			"enabled":  rule.Enabled,
		},
	})

	return rule, nil
}

// DeleteRule deletes a firewall rule
func (s *Service) DeleteRule(ctx context.Context, token string, tenantID, id uuid.UUID) error {
	// Get rule name for audit log
	rule, _ := s.repo.GetRuleByID(tenantID, id)
	ruleName := ""
	if rule != nil {
		ruleName = rule.Name
	}

	if err := s.repo.DeleteRule(tenantID, id); err != nil {
		return err
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallRuleDeleted,
		tenantID,
		id.String(),
		nil,
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.rule.deleted",
		ResourceType: "firewall_rule",
		ResourceID:   id.String(),
		Details: map[string]interface{}{
			"name": ruleName,
		},
	})

	return nil
}

// BulkDeleteRules deletes multiple rules by IDs
func (s *Service) BulkDeleteRules(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) (int64, error) {
	return s.repo.BulkDeleteRules(tenantID, ids)
}

// CountRules returns the total count of rules
func (s *Service) CountRules(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return s.repo.CountRules(tenantID)
}

// ========================================
// Firewall Profiles
// ========================================

// CreateProfile creates a new firewall profile
func (s *Service) CreateProfile(ctx context.Context, token string, tenantID, userID uuid.UUID, input *FirewallProfileInput) (*FirewallProfile, error) {
	enabled := true
	isDefault := false
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if input.IsDefault != nil {
		isDefault = *input.IsDefault
	}

	// Default feature settings
	enableNAT := false
	enableConntrack := true
	allowLoopback := true
	allowEstablished := true
	allowICMPPing := true
	enableIPv6 := false
	if input.EnableNAT != nil {
		enableNAT = *input.EnableNAT
	}
	if input.EnableConntrack != nil {
		enableConntrack = *input.EnableConntrack
	}
	if input.AllowLoopback != nil {
		allowLoopback = *input.AllowLoopback
	}
	if input.AllowEstablished != nil {
		allowEstablished = *input.AllowEstablished
	}
	if input.AllowICMPPing != nil {
		allowICMPPing = *input.AllowICMPPing
	}
	if input.EnableIPv6 != nil {
		enableIPv6 = *input.EnableIPv6
	}

	// Default policies
	inputPolicy := "drop"
	outputPolicy := "accept"
	forwardPolicy := "drop"
	if input.InputPolicy != "" {
		inputPolicy = input.InputPolicy
	}
	if input.OutputPolicy != "" {
		outputPolicy = input.OutputPolicy
	}
	if input.ForwardPolicy != "" {
		forwardPolicy = input.ForwardPolicy
	}

	profile := &FirewallProfile{
		TenantID:         tenantID,
		Name:             input.Name,
		Description:      input.Description,
		IsDefault:        isDefault,
		Enabled:          enabled,
		InputPolicy:      inputPolicy,
		OutputPolicy:     outputPolicy,
		ForwardPolicy:    forwardPolicy,
		EnableNAT:        enableNAT,
		EnableConntrack:  enableConntrack,
		AllowLoopback:    allowLoopback,
		AllowEstablished: allowEstablished,
		AllowICMPPing:    allowICMPPing,
		EnableIPv6:       enableIPv6,
		CreatedBy:        userID,
	}

	if err := s.repo.CreateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Add rules if provided
	if len(input.RuleIDs) > 0 {
		ruleIDs := make([]uuid.UUID, 0, len(input.RuleIDs))
		for _, idStr := range input.RuleIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				ruleIDs = append(ruleIDs, id)
			}
		}
		if len(ruleIDs) > 0 {
			s.repo.AddRulesToProfile(tenantID, profile.ID, ruleIDs)
		}
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallProfileCreated,
		tenantID,
		profile.ID.String(),
		map[string]interface{}{
			"name":      profile.Name,
			"isDefault": profile.IsDefault,
		},
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.profile.created",
		ResourceType: "firewall_profile",
		ResourceID:   profile.ID.String(),
		Details: map[string]interface{}{
			"name":      profile.Name,
			"isDefault": profile.IsDefault,
			"enabled":   profile.Enabled,
			"ruleCount": len(input.RuleIDs),
		},
	})

	return profile, nil
}

// GetProfile retrieves a profile by ID
func (s *Service) GetProfile(ctx context.Context, tenantID, id uuid.UUID) (*FirewallProfile, error) {
	return s.repo.GetProfileByID(tenantID, id)
}

// GetProfileWithRules retrieves a profile with its rules
func (s *Service) GetProfileWithRules(ctx context.Context, tenantID, id uuid.UUID) (*FirewallProfile, error) {
	return s.repo.GetProfileByIDWithRules(tenantID, id)
}

// ListProfiles retrieves all profiles for a tenant
func (s *Service) ListProfiles(ctx context.Context, tenantID uuid.UUID, filter *FirewallProfileFilter, limit, offset int) ([]FirewallProfile, int64, error) {
	p := pagination.Normalize(limit, offset)
	return s.repo.ListProfiles(tenantID, filter, p.Limit, p.Offset)
}

// UpdateProfile updates a firewall profile
func (s *Service) UpdateProfile(ctx context.Context, token string, tenantID, id uuid.UUID, input *FirewallProfileInput) (*FirewallProfile, error) {
	profile, err := s.repo.GetProfileByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		profile.Name = input.Name
	}
	if input.Description != "" {
		profile.Description = input.Description
	}
	if input.IsDefault != nil {
		profile.IsDefault = *input.IsDefault
	}
	if input.Enabled != nil {
		profile.Enabled = *input.Enabled
	}
	// Default policies
	if input.InputPolicy != "" {
		profile.InputPolicy = input.InputPolicy
	}
	if input.OutputPolicy != "" {
		profile.OutputPolicy = input.OutputPolicy
	}
	if input.ForwardPolicy != "" {
		profile.ForwardPolicy = input.ForwardPolicy
	}
	// Features
	if input.EnableNAT != nil {
		profile.EnableNAT = *input.EnableNAT
	}
	if input.EnableConntrack != nil {
		profile.EnableConntrack = *input.EnableConntrack
	}
	if input.AllowLoopback != nil {
		profile.AllowLoopback = *input.AllowLoopback
	}
	if input.AllowEstablished != nil {
		profile.AllowEstablished = *input.AllowEstablished
	}
	if input.AllowICMPPing != nil {
		profile.AllowICMPPing = *input.AllowICMPPing
	}
	if input.EnableIPv6 != nil {
		profile.EnableIPv6 = *input.EnableIPv6
	}

	if err := s.repo.UpdateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Update rules if provided
	if input.RuleIDs != nil {
		ruleIDs := make([]uuid.UUID, 0, len(input.RuleIDs))
		for _, idStr := range input.RuleIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				ruleIDs = append(ruleIDs, id)
			}
		}
		s.repo.SetProfileRules(tenantID, profile.ID, ruleIDs)
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallProfileUpdated,
		tenantID,
		profile.ID.String(),
		map[string]interface{}{
			"name":      profile.Name,
			"isDefault": profile.IsDefault,
		},
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.profile.updated",
		ResourceType: "firewall_profile",
		ResourceID:   profile.ID.String(),
		Details: map[string]interface{}{
			"name":      profile.Name,
			"isDefault": profile.IsDefault,
			"enabled":   profile.Enabled,
		},
	})

	return profile, nil
}

// DeleteProfile deletes a firewall profile
func (s *Service) DeleteProfile(ctx context.Context, token string, tenantID, id uuid.UUID) error {
	// Get profile name for audit log
	profile, _ := s.repo.GetProfileByID(tenantID, id)
	profileName := ""
	if profile != nil {
		profileName = profile.Name
	}

	if err := s.repo.DeleteProfile(tenantID, id); err != nil {
		return err
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallProfileDeleted,
		tenantID,
		id.String(),
		nil,
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.profile.deleted",
		ResourceType: "firewall_profile",
		ResourceID:   id.String(),
		Details: map[string]interface{}{
			"name": profileName,
		},
	})

	return nil
}

// AddRulesToProfile adds rules to a profile
func (s *Service) AddRulesToProfile(ctx context.Context, tenantID, profileID uuid.UUID, ruleIDs []uuid.UUID) error {
	// Verify profile exists and belongs to tenant
	if _, err := s.repo.GetProfileByID(tenantID, profileID); err != nil {
		return err
	}
	// AddRulesToProfile validates that rules also belong to this tenant
	return s.repo.AddRulesToProfile(tenantID, profileID, ruleIDs)
}

// RemoveRulesFromProfile removes rules from a profile
func (s *Service) RemoveRulesFromProfile(ctx context.Context, tenantID, profileID uuid.UUID, ruleIDs []uuid.UUID) error {
	// Verify profile exists
	if _, err := s.repo.GetProfileByID(tenantID, profileID); err != nil {
		return err
	}
	return s.repo.RemoveRulesFromProfile(profileID, ruleIDs)
}

// CountProfiles returns the total count of profiles
func (s *Service) CountProfiles(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return s.repo.CountProfiles(tenantID)
}

// ========================================
// Firewall Templates
// ========================================

// CreateTemplate creates a new firewall template
func (s *Service) CreateTemplate(ctx context.Context, token string, tenantID, userID uuid.UUID, input *FirewallTemplateInput) (*FirewallTemplate, error) {
	rulesJSON, err := json.Marshal(input.Rules)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize rules: %w", err)
	}

	template := &FirewallTemplate{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		IsBuiltIn:   false,
		RulesJSON:   string(rulesJSON),
		CreatedBy:   userID,
	}

	if template.Category == "" {
		template.Category = TemplateCategoryCustom
	}

	if err := s.repo.CreateTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallTemplateCreated,
		tenantID,
		template.ID.String(),
		map[string]interface{}{
			"name":     template.Name,
			"category": template.Category,
		},
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.template.created",
		ResourceType: "firewall_template",
		ResourceID:   template.ID.String(),
		Details: map[string]interface{}{
			"name":      template.Name,
			"category":  template.Category,
			"ruleCount": len(input.Rules),
		},
	})

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *Service) GetTemplate(ctx context.Context, tenantID, id uuid.UUID) (*FirewallTemplate, error) {
	return s.repo.GetTemplateByID(tenantID, id)
}

// ListTemplates retrieves all templates for a tenant
func (s *Service) ListTemplates(ctx context.Context, tenantID uuid.UUID, filter *FirewallTemplateFilter, limit, offset int) ([]FirewallTemplate, int64, error) {
	p := pagination.Normalize(limit, offset)
	return s.repo.ListTemplates(tenantID, filter, p.Limit, p.Offset)
}

// UpdateTemplate updates a firewall template
func (s *Service) UpdateTemplate(ctx context.Context, token string, tenantID, id uuid.UUID, input *FirewallTemplateInput) (*FirewallTemplate, error) {
	template, err := s.repo.GetTemplateByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	// Don't allow updating built-in templates
	if template.IsBuiltIn {
		return nil, fmt.Errorf("cannot update built-in template")
	}

	if input.Name != "" {
		template.Name = input.Name
	}
	if input.Description != "" {
		template.Description = input.Description
	}
	if input.Category != "" {
		template.Category = input.Category
	}
	if input.Rules != nil {
		rulesJSON, err := json.Marshal(input.Rules)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize rules: %w", err)
		}
		template.RulesJSON = string(rulesJSON)
	}

	if err := s.repo.UpdateTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallTemplateUpdated,
		tenantID,
		template.ID.String(),
		map[string]interface{}{
			"name":     template.Name,
			"category": template.Category,
		},
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.template.updated",
		ResourceType: "firewall_template",
		ResourceID:   template.ID.String(),
		Details: map[string]interface{}{
			"name":     template.Name,
			"category": template.Category,
		},
	})

	return template, nil
}

// DeleteTemplate deletes a firewall template
func (s *Service) DeleteTemplate(ctx context.Context, token string, tenantID, id uuid.UUID) error {
	// Get template name for audit log
	template, _ := s.repo.GetTemplateByID(tenantID, id)
	templateName := ""
	if template != nil {
		templateName = template.Name
	}

	if err := s.repo.DeleteTemplate(tenantID, id); err != nil {
		return err
	}

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallTemplateDeleted,
		tenantID,
		id.String(),
		nil,
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.template.deleted",
		ResourceType: "firewall_template",
		ResourceID:   id.String(),
		Details: map[string]interface{}{
			"name": templateName,
		},
	})

	return nil
}

// ApplyTemplateToProfile applies a template's rules to a profile
func (s *Service) ApplyTemplateToProfile(ctx context.Context, token string, tenantID, userID, templateID, profileID uuid.UUID) error {
	template, err := s.repo.GetTemplateByID(tenantID, templateID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	profile, err := s.repo.GetProfileByID(tenantID, profileID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	// Parse template rules
	rules, err := s.repo.GetTemplateRules(template)
	if err != nil {
		return fmt.Errorf("failed to parse template rules: %w", err)
	}

	// Create rules from template and add to profile
	ruleIDs := make([]uuid.UUID, 0, len(rules))
	for _, ruleDef := range rules {
		rule := &FirewallRule{
			TenantID:    tenantID,
			Name:        ruleDef.Name,
			Description: ruleDef.Description,
			Chain:       ruleDef.Chain,
			Priority:    ruleDef.Priority,
			Protocol:    ruleDef.Protocol,
			SourceIP:    ruleDef.SourceIP,
			SourcePort:  ruleDef.SourcePort,
			DestIP:      ruleDef.DestIP,
			DestPort:    ruleDef.DestPort,
			Action:      ruleDef.Action,
			Comment:     ruleDef.Comment,
			Enabled:     true,
			CreatedBy:   userID,
		}
		if err := s.repo.CreateRule(rule); err != nil {
			continue // Skip failed rules
		}
		ruleIDs = append(ruleIDs, rule.ID)
	}

	// Add rules to profile (tenantID for validation)
	if len(ruleIDs) > 0 {
		if err := s.repo.AddRulesToProfile(tenantID, profile.ID, ruleIDs); err != nil {
			return fmt.Errorf("failed to add rules to profile: %w", err)
		}
	}

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.template.applied",
		ResourceType: "firewall_profile",
		ResourceID:   profile.ID.String(),
		Details: map[string]interface{}{
			"templateId":   template.ID.String(),
			"templateName": template.Name,
			"profileName":  profile.Name,
			"rulesCreated": len(ruleIDs),
		},
	})

	return nil
}

// CountTemplates returns the total count of templates
func (s *Service) CountTemplates(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return s.repo.CountTemplates(tenantID)
}

// ========================================
// Firewall Deployments
// ========================================

// DeployProfile deploys a profile to an agent using nftables_apply playbook
func (s *Service) DeployProfile(ctx context.Context, token string, tenantID, userID uuid.UUID, input *DeploymentInput) (*FirewallDeployment, error) {
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profileId: %w", err)
	}

	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agentId: %w", err)
	}

	// Validate agent capability (nftables)
	if err := s.client.ValidateAgentCapability(ctx, token, agentID, "nftables"); err != nil {
		return nil, fmt.Errorf("agent capability validation failed: %w", err)
	}

	// Get profile with rules
	profile, err := s.repo.GetProfileByIDWithRules(tenantID, profileID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// Get agent name from csd-core
	agentName := "Unknown"
	if agent, err := s.client.GetAgent(ctx, token, agentID); err == nil && agent != nil {
		agentName = agent.Name
	}

	// Create snapshot of rules
	rulesSnapshot, _ := json.Marshal(profile.Rules)

	// Determine status based on dry-run mode
	status := DeploymentStatusPending
	if input.DryRun {
		status = DeploymentStatusApplied // Dry-run is instant validation
	}

	deployment := &FirewallDeployment{
		TenantID:      tenantID,
		ProfileID:     &profileID,
		AgentID:       agentID,
		AgentName:     agentName,
		Action:        DeploymentActionApply,
		Status:        status,
		RulesSnapshot: string(rulesSnapshot),
		CreatedBy:     userID,
	}

	if err := s.repo.CreateDeployment(deployment); err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.deployment.initiated",
		ResourceType: "firewall_deployment",
		ResourceID:   deployment.ID.String(),
		Details: map[string]interface{}{
			"profileId":   profile.ID.String(),
			"profileName": profile.Name,
			"agentId":     agentID.String(),
			"agentName":   agentName,
			"dryRun":      input.DryRun,
			"ruleCount":   len(profile.Rules),
		},
	})

	// For dry-run mode, just validate and return
	if input.DryRun {
		nftConfig := s.generateNftablesConfig(profile.Rules)
		s.repo.UpdateDeploymentStatus(deployment.ID, DeploymentStatusApplied,
			"Dry-run validation successful. Configuration is valid.",
			nftConfig)
		return deployment, nil
	}

	// Start async deployment
	go s.runDeployment(deployment.ID, tenantID, token, profile, agentID)

	return deployment, nil
}

// runDeployment executes the deployment in background
func (s *Service) runDeployment(deploymentID, tenantID uuid.UUID, token string, profile *FirewallProfile, agentID uuid.UUID) {
	// Use timeout to prevent goroutine leaks
	timeout := 5 * time.Minute
	if cfg := config.GetConfig(); cfg != nil && cfg.Limits.FirewallDeploymentTimeout > 0 {
		timeout = time.Duration(cfg.Limits.FirewallDeploymentTimeout) * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s.repo.UpdateDeploymentStatus(deploymentID, DeploymentStatusDeploying, "Applying firewall rules...", "")

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallDeployStarted,
		tenantID,
		deploymentID.String(),
		map[string]interface{}{
			"profileId": profile.ID.String(),
			"agentId":   agentID.String(),
		},
	))

	// Generate nftables configuration from profile (includes ct state, loopback, NAT)
	nftConfig := s.generateNftablesConfigForProfile(profile)

	// Store backup of current configuration via csd-core Artifacts
	backupKey := fmt.Sprintf("firewall-backup-%s-%s", agentID.String(), time.Now().Format("20060102-150405"))
	backupData := map[string]interface{}{
		"profile_id":   profile.ID.String(),
		"profile_name": profile.Name,
		"rules":        profile.Rules,
		"config":       nftConfig,
	}
	backupJSON, _ := json.Marshal(backupData)
	if err := s.client.CreateArtifact(ctx, token, tenantID, backupKey, "firewall-backup", string(backupJSON)); err != nil {
		// Log but don't fail - backup is best effort
		s.repo.UpdateDeploymentStatus(deploymentID, DeploymentStatusDeploying,
			fmt.Sprintf("Backup creation failed (continuing): %s", err.Error()), "")
	}

	// Execute nftables task via csd-core using config_content
	// This deploys the complete nftables configuration file
	execution, err := s.client.ExecuteTask(ctx, token, &csdcore.ExecuteTaskInput{
		AgentID: agentID,
		Task: csdcore.TaskInput{
			Type: "nftables",
			Name: fmt.Sprintf("deploy-profile-%s", profile.Name),
			Config: map[string]interface{}{
				"config_content": nftConfig,
			},
		},
		Wait:    true,
		Timeout: 120,
	})
	if err != nil {
		s.repo.UpdateDeploymentStatus(deploymentID, DeploymentStatusError, "Failed to execute task: "+err.Error(), "")
		events.GetEventBus().PublishAsync(events.NewEvent(
			events.EventFirewallDeployFailed,
			tenantID,
			deploymentID.String(),
			map[string]interface{}{"error": err.Error()},
		))

		// Audit log for failure
		s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
			Action:       "firewall.deployment.failed",
			ResourceType: "firewall_deployment",
			ResourceID:   deploymentID.String(),
			Details: map[string]interface{}{
				"profileId": profile.ID.String(),
				"agentId":   agentID.String(),
				"error":     err.Error(),
			},
		})
		return
	}

	if execution.Status != "SUCCESS" {
		output := ""
		if execution.Output != nil {
			if str, ok := execution.Output.(string); ok {
				output = str
			}
		}
		s.repo.UpdateDeploymentStatus(deploymentID, DeploymentStatusError, "Task failed: "+execution.Error, output)
		events.GetEventBus().PublishAsync(events.NewEvent(
			events.EventFirewallDeployFailed,
			tenantID,
			deploymentID.String(),
			map[string]interface{}{"error": execution.Error},
		))

		// Audit log for failure
		s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
			Action:       "firewall.deployment.failed",
			ResourceType: "firewall_deployment",
			ResourceID:   deploymentID.String(),
			Details: map[string]interface{}{
				"profileId": profile.ID.String(),
				"agentId":   agentID.String(),
				"error":     execution.Error,
			},
		})
		return
	}

	output := ""
	if execution.Output != nil {
		if str, ok := execution.Output.(string); ok {
			output = str
		}
	}
	s.repo.UpdateDeploymentStatus(deploymentID, DeploymentStatusApplied, "Firewall rules applied successfully", output)
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallDeployCompleted,
		tenantID,
		deploymentID.String(),
		map[string]interface{}{
			"profileId": profile.ID.String(),
			"agentId":   agentID.String(),
		},
	))

	// Audit log for success
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.deployment.completed",
		ResourceType: "firewall_deployment",
		ResourceID:   deploymentID.String(),
		Details: map[string]interface{}{
			"profileId": profile.ID.String(),
			"agentId":   agentID.String(),
			"backupKey": backupKey,
		},
	})
}

// generateNftablesConfigForProfile generates complete nftables configuration from a profile
func (s *Service) generateNftablesConfigForProfile(profile *FirewallProfile) string {
	var config strings.Builder
	// Pre-allocate reasonable capacity (reduces reallocations)
	config.Grow(4096)

	config.WriteString("#!/usr/sbin/nft -f\n\n")
	config.WriteString("# Generated by CSD-Pilote Security Module\n")
	fmt.Fprintf(&config, "# Profile: %s\n", profile.Name)
	fmt.Fprintf(&config, "# Generated at: %s\n\n", time.Now().Format(time.RFC3339))
	config.WriteString("flush ruleset\n\n")

	// Determine family (inet = IPv4+IPv6, ip = IPv4 only)
	family := "inet"
	if !profile.EnableIPv6 {
		family = "ip"
	}

	// Filter table
	fmt.Fprintf(&config, "table %s filter {\n", family)

	// Group rules by chain
	chainRules := make(map[RuleChain][]FirewallRule)
	for _, rule := range profile.Rules {
		if rule.Enabled {
			chainRules[rule.Chain] = append(chainRules[rule.Chain], rule)
		}
	}

	// Generate filter chains
	chains := []struct {
		name       RuleChain
		nftName    string
		hookType   string
		policyFunc func() string
	}{
		{RuleChainInput, "input", "input", func() string { return profile.InputPolicy }},
		{RuleChainOutput, "output", "output", func() string { return profile.OutputPolicy }},
		{RuleChainForward, "forward", "forward", func() string { return profile.ForwardPolicy }},
	}

	for _, chain := range chains {
		policy := chain.policyFunc()
		if policy == "" {
			policy = "drop"
		}
		fmt.Fprintf(&config, "    chain %s {\n", chain.nftName)
		fmt.Fprintf(&config, "        type filter hook %s priority 0; policy %s;\n\n", chain.hookType, policy)

		// Add base rules based on profile settings
		if chain.name == RuleChainInput {
			// Loopback rule
			if profile.AllowLoopback {
				config.WriteString("        # Allow loopback traffic\n")
				config.WriteString("        iif lo accept\n\n")
			}

			// Connection tracking
			if profile.AllowEstablished {
				config.WriteString("        # Allow established and related connections\n")
				config.WriteString("        ct state established,related accept\n")
				config.WriteString("        ct state invalid drop\n\n")
			}

			// ICMP ping
			if profile.AllowICMPPing {
				config.WriteString("        # Allow ICMP ping\n")
				if family == "inet" {
					config.WriteString("        ip protocol icmp icmp type echo-request accept\n")
					config.WriteString("        ip6 nexthdr icmpv6 icmpv6 type echo-request accept\n\n")
				} else {
					config.WriteString("        ip protocol icmp icmp type echo-request accept\n\n")
				}
			}
		}

		if chain.name == RuleChainOutput {
			// Loopback output
			if profile.AllowLoopback {
				config.WriteString("        # Allow loopback traffic\n")
				config.WriteString("        oif lo accept\n\n")
			}

			// Connection tracking for output
			if profile.AllowEstablished {
				config.WriteString("        # Allow established and related connections\n")
				config.WriteString("        ct state established,related accept\n\n")
			}
		}

		if chain.name == RuleChainForward {
			// Connection tracking for forward
			if profile.AllowEstablished {
				config.WriteString("        # Allow established and related connections\n")
				config.WriteString("        ct state established,related accept\n")
				config.WriteString("        ct state invalid drop\n\n")
			}
		}

		// Add user-defined rules
		for _, rule := range chainRules[chain.name] {
			config.WriteString("        ")
			config.WriteString(s.ruleToNft(rule))
			config.WriteByte('\n')
		}

		config.WriteString("    }\n\n")
	}

	config.WriteString("}\n\n")

	// NAT table (if enabled)
	if profile.EnableNAT {
		fmt.Fprintf(&config, "table %s nat {\n", family)

		// Prerouting chain (for DNAT)
		config.WriteString("    chain prerouting {\n")
		config.WriteString("        type nat hook prerouting priority dstnat;\n\n")
		for _, rule := range chainRules[RuleChainPrerouting] {
			config.WriteString("        ")
			config.WriteString(s.ruleToNft(rule))
			config.WriteByte('\n')
		}
		config.WriteString("    }\n\n")

		// Postrouting chain (for SNAT/MASQUERADE)
		config.WriteString("    chain postrouting {\n")
		config.WriteString("        type nat hook postrouting priority srcnat;\n\n")
		for _, rule := range chainRules[RuleChainPostrouting] {
			config.WriteString("        ")
			config.WriteString(s.ruleToNft(rule))
			config.WriteByte('\n')
		}
		config.WriteString("    }\n")

		config.WriteString("}\n")
	}

	return config.String()
}

// generateNftablesConfig generates nftables configuration from rules (legacy, for dry-run)
func (s *Service) generateNftablesConfig(rules []FirewallRule) string {
	// Create a temporary profile with default settings
	profile := &FirewallProfile{
		Name:             "Dry-Run Profile",
		InputPolicy:      "drop",
		OutputPolicy:     "accept",
		ForwardPolicy:    "drop",
		AllowLoopback:    true,
		AllowEstablished: true,
		AllowICMPPing:    true,
		EnableNAT:        false,
		EnableIPv6:       false,
		Rules:            rules,
	}
	return s.generateNftablesConfigForProfile(profile)
}

// ruleToNft converts a FirewallRule to nftables syntax
func (s *Service) ruleToNft(rule FirewallRule) string {
	// If raw expression is provided, use it directly
	if rule.RuleExpr != "" {
		return fmt.Sprintf("%s # %s", rule.RuleExpr, rule.Name)
	}

	var parts []string

	// Interface matching
	if rule.InInterface != "" {
		parts = append(parts, fmt.Sprintf("iif %s", rule.InInterface))
	}
	if rule.OutInterface != "" {
		parts = append(parts, fmt.Sprintf("oif %s", rule.OutInterface))
	}

	// Connection tracking state
	if rule.CTState != "" {
		parts = append(parts, fmt.Sprintf("ct state %s", strings.ToLower(rule.CTState)))
	}

	// Protocol
	if rule.Protocol != "" && rule.Protocol != RuleProtocolAll {
		proto := strings.ToLower(string(rule.Protocol))
		parts = append(parts, fmt.Sprintf("ip protocol %s", proto))
	}

	// Source IP
	if rule.SourceIP != "" {
		parts = append(parts, fmt.Sprintf("ip saddr %s", rule.SourceIP))
	}

	// Destination IP
	if rule.DestIP != "" {
		parts = append(parts, fmt.Sprintf("ip daddr %s", rule.DestIP))
	}

	// Source port (requires TCP or UDP)
	if rule.SourcePort != "" {
		proto := strings.ToLower(string(rule.Protocol))
		if proto == "tcp" || proto == "udp" {
			parts = append(parts, fmt.Sprintf("%s sport %s", proto, rule.SourcePort))
		} else {
			parts = append(parts, fmt.Sprintf("th sport %s", rule.SourcePort))
		}
	}

	// Destination port (requires TCP or UDP)
	if rule.DestPort != "" {
		proto := strings.ToLower(string(rule.Protocol))
		if proto == "tcp" || proto == "udp" {
			parts = append(parts, fmt.Sprintf("%s dport %s", proto, rule.DestPort))
		} else {
			parts = append(parts, fmt.Sprintf("th dport %s", rule.DestPort))
		}
	}

	// Rate limiting
	if rule.RateLimit != "" {
		limitExpr := fmt.Sprintf("limit rate %s", rule.RateLimit)
		if rule.RateBurst > 0 {
			limitExpr += fmt.Sprintf(" burst %d packets", rule.RateBurst)
		}
		parts = append(parts, limitExpr)
	}

	// Action
	action := s.actionToNft(rule)
	parts = append(parts, action)

	// Comment
	if rule.Comment != "" {
		// Escape quotes in comment
		comment := strings.ReplaceAll(rule.Comment, "\"", "\\\"")
		parts = append(parts, fmt.Sprintf("comment \"%s\"", comment))
	}

	return fmt.Sprintf("%s # %s", joinParts(parts), rule.Name)
}

// actionToNft converts a rule action to nftables syntax
func (s *Service) actionToNft(rule FirewallRule) string {
	switch rule.Action {
	case RuleActionAccept:
		return "accept"
	case RuleActionDrop:
		return "drop"
	case RuleActionReject:
		return "reject"
	case RuleActionLog:
		logExpr := "log"
		if rule.LogPrefix != "" {
			logExpr += fmt.Sprintf(" prefix \"%s\"", rule.LogPrefix)
		}
		if rule.LogLevel != "" {
			logExpr += fmt.Sprintf(" level %s", rule.LogLevel)
		}
		return logExpr
	case RuleActionMasquerade:
		return "masquerade"
	case RuleActionSnat:
		if rule.NatToAddr != "" {
			return fmt.Sprintf("snat to %s", rule.NatToAddr)
		}
		return "snat"
	case RuleActionDnat:
		target := ""
		if rule.NatToAddr != "" {
			target = rule.NatToAddr
			if rule.NatToPort != "" {
				target += ":" + rule.NatToPort
			}
			return fmt.Sprintf("dnat to %s", target)
		}
		return "dnat"
	case RuleActionRedirect:
		if rule.NatToPort != "" {
			return fmt.Sprintf("redirect to :%s", rule.NatToPort)
		}
		return "redirect"
	default:
		return "accept"
	}
}

func joinParts(parts []string) string {
	return strings.Join(parts, " ")
}

// RollbackDeployment rolls back a deployment using nftables_rollback playbook
func (s *Service) RollbackDeployment(ctx context.Context, token string, tenantID, userID, deploymentID uuid.UUID) (*FirewallDeployment, error) {
	originalDeployment, err := s.repo.GetDeploymentByID(tenantID, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("deployment not found: %w", err)
	}

	// Create rollback deployment record
	rollback := &FirewallDeployment{
		TenantID:      tenantID,
		ProfileID:     originalDeployment.ProfileID,
		AgentID:       originalDeployment.AgentID,
		AgentName:     originalDeployment.AgentName,
		Action:        DeploymentActionRollback,
		Status:        DeploymentStatusPending,
		CreatedBy:     userID,
	}

	if err := s.repo.CreateDeployment(rollback); err != nil {
		return nil, fmt.Errorf("failed to create rollback record: %w", err)
	}

	// Execute rollback asynchronously
	go s.runRollback(rollback.ID, tenantID, token, originalDeployment.AgentID)

	return rollback, nil
}

// runRollback executes the rollback in background
func (s *Service) runRollback(rollbackID, tenantID uuid.UUID, token string, agentID uuid.UUID) {
	// Use timeout to prevent goroutine leaks (2 minutes max for rollback)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	s.repo.UpdateDeploymentStatus(rollbackID, DeploymentStatusDeploying, "Rolling back firewall rules...", "")

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallRollbackStarted,
		tenantID,
		rollbackID.String(),
		map[string]interface{}{"agentId": agentID.String()},
	))

	execution, err := s.client.ExecuteTask(ctx, token, &csdcore.ExecuteTaskInput{
		AgentID: agentID,
		Task: csdcore.TaskInput{
			Type: "nftables",
			Name: "nftables-rollback",
			Config: map[string]interface{}{
				"action": "rollback",
			},
		},
		Wait:    true,
		Timeout: 60,
	})
	if err != nil {
		s.repo.UpdateDeploymentStatus(rollbackID, DeploymentStatusError, "Failed to execute rollback: "+err.Error(), "")
		return
	}

	output := ""
	if execution.Output != nil {
		if s, ok := execution.Output.(string); ok {
			output = s
		}
	}

	if execution.Status != "SUCCESS" {
		s.repo.UpdateDeploymentStatus(rollbackID, DeploymentStatusError, "Rollback failed: "+execution.Error, output)
		return
	}

	s.repo.UpdateDeploymentStatus(rollbackID, DeploymentStatusRolledBack, "Firewall rules rolled back successfully", output)
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallRollbackCompleted,
		tenantID,
		rollbackID.String(),
		map[string]interface{}{"agentId": agentID.String()},
	))
}

// AuditDeployment audits the current firewall state on an agent
func (s *Service) AuditDeployment(ctx context.Context, token string, tenantID, userID uuid.UUID, agentID uuid.UUID) (*FirewallDeployment, error) {
	// Get agent name
	agentName := "Unknown"
	if agent, err := s.client.GetAgent(ctx, token, agentID); err == nil && agent != nil {
		agentName = agent.Name
	}

	audit := &FirewallDeployment{
		TenantID:  tenantID,
		AgentID:   agentID,
		AgentName: agentName,
		Action:    DeploymentActionAudit,
		Status:    DeploymentStatusPending,
		CreatedBy: userID,
	}

	if err := s.repo.CreateDeployment(audit); err != nil {
		return nil, fmt.Errorf("failed to create audit record: %w", err)
	}

	// Execute audit asynchronously
	go s.runAudit(audit.ID, tenantID, token, agentID)

	return audit, nil
}

// runAudit executes the audit in background
func (s *Service) runAudit(auditID, tenantID uuid.UUID, token string, agentID uuid.UUID) {
	// Use timeout to prevent goroutine leaks (2 minutes max for audit)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	s.repo.UpdateDeploymentStatus(auditID, DeploymentStatusDeploying, "Auditing firewall rules...", "")

	execution, err := s.client.ExecuteTask(ctx, token, &csdcore.ExecuteTaskInput{
		AgentID: agentID,
		Task: csdcore.TaskInput{
			Type: "nftables",
			Name: "nftables-audit",
			Config: map[string]interface{}{
				"action": "audit",
			},
		},
		Wait:    true,
		Timeout: 60,
	})
	if err != nil {
		s.repo.UpdateDeploymentStatus(auditID, DeploymentStatusError, "Failed to execute audit: "+err.Error(), "")
		return
	}

	output := ""
	if execution.Output != nil {
		if s, ok := execution.Output.(string); ok {
			output = s
		}
	}

	if execution.Status != "SUCCESS" {
		s.repo.UpdateDeploymentStatus(auditID, DeploymentStatusError, "Audit failed: "+execution.Error, output)
		return
	}

	s.repo.UpdateDeploymentStatus(auditID, DeploymentStatusApplied, "Audit completed successfully", output)
}

// FlushRules flushes all firewall rules on an agent
func (s *Service) FlushRules(ctx context.Context, token string, tenantID, userID uuid.UUID, agentID uuid.UUID) (*FirewallDeployment, error) {
	// Get agent name
	agentName := "Unknown"
	if agent, err := s.client.GetAgent(ctx, token, agentID); err == nil && agent != nil {
		agentName = agent.Name
	}

	flush := &FirewallDeployment{
		TenantID:  tenantID,
		AgentID:   agentID,
		AgentName: agentName,
		Action:    DeploymentActionFlush,
		Status:    DeploymentStatusPending,
		CreatedBy: userID,
	}

	if err := s.repo.CreateDeployment(flush); err != nil {
		return nil, fmt.Errorf("failed to create flush record: %w", err)
	}

	// Execute flush asynchronously
	go s.runFlush(flush.ID, tenantID, token, agentID)

	return flush, nil
}

// runFlush executes the flush in background
func (s *Service) runFlush(flushID, tenantID uuid.UUID, token string, agentID uuid.UUID) {
	// Use timeout to prevent goroutine leaks (2 minutes max for flush)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	s.repo.UpdateDeploymentStatus(flushID, DeploymentStatusDeploying, "Flushing firewall rules...", "")

	execution, err := s.client.ExecuteTask(ctx, token, &csdcore.ExecuteTaskInput{
		AgentID: agentID,
		Task: csdcore.TaskInput{
			Type: "nftables",
			Name: "nftables-flush",
			Config: map[string]interface{}{
				"action":        "flush",
				"confirm_flush": true,
			},
		},
		Wait:    true,
		Timeout: 60,
	})
	if err != nil {
		s.repo.UpdateDeploymentStatus(flushID, DeploymentStatusError, "Failed to execute flush: "+err.Error(), "")
		return
	}

	output := ""
	if execution.Output != nil {
		if s, ok := execution.Output.(string); ok {
			output = s
		}
	}

	if execution.Status != "SUCCESS" {
		s.repo.UpdateDeploymentStatus(flushID, DeploymentStatusError, "Flush failed: "+execution.Error, output)
		return
	}

	s.repo.UpdateDeploymentStatus(flushID, DeploymentStatusApplied, "Firewall rules flushed successfully", output)
}

// GetDeployment retrieves a deployment by ID
func (s *Service) GetDeployment(ctx context.Context, tenantID, id uuid.UUID) (*FirewallDeployment, error) {
	return s.repo.GetDeploymentByID(tenantID, id)
}

// ListDeployments retrieves all deployments for a tenant
func (s *Service) ListDeployments(ctx context.Context, tenantID uuid.UUID, filter *FirewallDeploymentFilter, limit, offset int) ([]FirewallDeployment, int64, error) {
	p := pagination.Normalize(limit, offset)
	return s.repo.ListDeployments(tenantID, filter, p.Limit, p.Offset)
}

// CountDeployments returns the total count of deployments
func (s *Service) CountDeployments(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return s.repo.CountDeployments(tenantID)
}

// ========================================
// Import/Export Functionality
// ========================================

// ExportProfile exports a profile with its rules to JSON format
func (s *Service) ExportProfile(ctx context.Context, token string, tenantID, profileID uuid.UUID) (*ProfileExport, error) {
	profile, err := s.repo.GetProfileByIDWithRules(tenantID, profileID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// Convert rules to template rule definitions
	rules := make([]TemplateRuleDefinition, 0, len(profile.Rules))
	for _, rule := range profile.Rules {
		rules = append(rules, TemplateRuleDefinition{
			Name:        rule.Name,
			Description: rule.Description,
			Chain:       rule.Chain,
			Priority:    rule.Priority,
			Protocol:    rule.Protocol,
			SourceIP:    rule.SourceIP,
			SourcePort:  rule.SourcePort,
			DestIP:      rule.DestIP,
			DestPort:    rule.DestPort,
			Action:      rule.Action,
			Comment:     rule.Comment,
		})
	}

	export := &ProfileExport{
		Name:        profile.Name,
		Description: profile.Description,
		Rules:       rules,
		ExportedAt:  time.Now().Format(time.RFC3339),
	}

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.profile.exported",
		ResourceType: "firewall_profile",
		ResourceID:   profile.ID.String(),
		Details: map[string]interface{}{
			"name":      profile.Name,
			"ruleCount": len(rules),
		},
	})

	return export, nil
}

// ImportProfile imports a profile from JSON format
func (s *Service) ImportProfile(ctx context.Context, token string, tenantID, userID uuid.UUID, input *ProfileImportInput) (*FirewallProfile, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("profile name is required")
	}

	// Create the profile
	profile := &FirewallProfile{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		IsDefault:   false,
		Enabled:     true,
		CreatedBy:   userID,
	}

	if err := s.repo.CreateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Create rules from import and add to profile
	ruleIDs := make([]uuid.UUID, 0, len(input.Rules))
	for _, ruleDef := range input.Rules {
		rule := &FirewallRule{
			TenantID:    tenantID,
			Name:        ruleDef.Name,
			Description: ruleDef.Description,
			Chain:       ruleDef.Chain,
			Priority:    ruleDef.Priority,
			Protocol:    ruleDef.Protocol,
			SourceIP:    ruleDef.SourceIP,
			SourcePort:  ruleDef.SourcePort,
			DestIP:      ruleDef.DestIP,
			DestPort:    ruleDef.DestPort,
			Action:      ruleDef.Action,
			Comment:     ruleDef.Comment,
			Enabled:     true,
			CreatedBy:   userID,
		}
		if err := s.repo.CreateRule(rule); err != nil {
			continue // Skip failed rules
		}
		ruleIDs = append(ruleIDs, rule.ID)
	}

	// Add rules to profile (tenantID for validation)
	if len(ruleIDs) > 0 {
		if err := s.repo.AddRulesToProfile(tenantID, profile.ID, ruleIDs); err != nil {
			return nil, fmt.Errorf("failed to add rules to profile: %w", err)
		}
	}

	// Reload profile with rules
	profile, _ = s.repo.GetProfileByIDWithRules(tenantID, profile.ID)

	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventFirewallProfileCreated,
		tenantID,
		profile.ID.String(),
		map[string]interface{}{
			"name":     profile.Name,
			"imported": true,
		},
	))

	// Audit logging
	s.client.LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "firewall.profile.imported",
		ResourceType: "firewall_profile",
		ResourceID:   profile.ID.String(),
		Details: map[string]interface{}{
			"name":         profile.Name,
			"rulesCreated": len(ruleIDs),
		},
	})

	return profile, nil
}
