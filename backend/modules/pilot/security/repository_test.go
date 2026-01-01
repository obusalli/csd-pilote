package security

import (
	"testing"

	"github.com/google/uuid"
)

func TestFirewallRuleModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		rule := FirewallRule{}
		if rule.TableName() != "firewall_rules" {
			t.Errorf("Expected table name 'firewall_rules', got '%s'", rule.TableName())
		}
	})

	t.Run("rule chain values", func(t *testing.T) {
		validChains := []RuleChain{
			RuleChainInput,
			RuleChainOutput,
			RuleChainForward,
		}
		for _, chain := range validChains {
			if chain == "" {
				t.Error("Chain should not be empty")
			}
		}
	})

	t.Run("rule action values", func(t *testing.T) {
		validActions := []RuleAction{
			RuleActionAccept,
			RuleActionDrop,
			RuleActionReject,
		}
		for _, action := range validActions {
			if action == "" {
				t.Error("Action should not be empty")
			}
		}
	})

	t.Run("rule protocol values", func(t *testing.T) {
		validProtocols := []RuleProtocol{
			RuleProtocolTCP,
			RuleProtocolUDP,
			RuleProtocolICMP,
			RuleProtocolAll,
		}
		for _, protocol := range validProtocols {
			if protocol == "" {
				t.Error("Protocol should not be empty")
			}
		}
	})
}

func TestFirewallProfileModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		profile := FirewallProfile{}
		if profile.TableName() != "firewall_profiles" {
			t.Errorf("Expected table name 'firewall_profiles', got '%s'", profile.TableName())
		}
	})
}

func TestFirewallTemplateModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		template := FirewallTemplate{}
		if template.TableName() != "firewall_templates" {
			t.Errorf("Expected table name 'firewall_templates', got '%s'", template.TableName())
		}
	})

	t.Run("template category values", func(t *testing.T) {
		validCategories := []TemplateCategory{
			TemplateCategoryWebServer,
			TemplateCategoryDatabase,
			TemplateCategoryBastion,
			TemplateCategoryGateway,
			TemplateCategoryCustom,
		}
		for _, category := range validCategories {
			if category == "" {
				t.Error("Category should not be empty")
			}
		}
	})
}

func TestFirewallDeploymentModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		deployment := FirewallDeployment{}
		if deployment.TableName() != "firewall_deployments" {
			t.Errorf("Expected table name 'firewall_deployments', got '%s'", deployment.TableName())
		}
	})

	t.Run("deployment status values", func(t *testing.T) {
		validStatuses := []DeploymentStatus{
			DeploymentStatusPending,
			DeploymentStatusDeploying,
			DeploymentStatusApplied,
			DeploymentStatusRolledBack,
			DeploymentStatusError,
		}
		for _, status := range validStatuses {
			if status == "" {
				t.Error("Status should not be empty")
			}
		}
	})

	t.Run("deployment action values", func(t *testing.T) {
		validActions := []DeploymentAction{
			DeploymentActionApply,
			DeploymentActionRollback,
			DeploymentActionAudit,
			DeploymentActionFlush,
		}
		for _, action := range validActions {
			if action == "" {
				t.Error("Action should not be empty")
			}
		}
	})
}

func TestFirewallRuleInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		enabled := true
		input := FirewallRuleInput{
			Name:        "allow-ssh",
			Description: "Allow SSH connections",
			Chain:       RuleChainInput,
			Protocol:    RuleProtocolTCP,
			DestPort:    "22",
			Action:      RuleActionAccept,
			Priority:    100,
			Enabled:     &enabled,
		}
		if input.Name != "allow-ssh" {
			t.Error("Expected name to be 'allow-ssh'")
		}
		if input.Chain != RuleChainInput {
			t.Error("Expected chain to be INPUT")
		}
		if input.Protocol != RuleProtocolTCP {
			t.Error("Expected protocol to be tcp")
		}
		if input.DestPort != "22" {
			t.Error("Expected dest port to be '22'")
		}
		if input.Action != RuleActionAccept {
			t.Error("Expected action to be ACCEPT")
		}
	})
}

func TestFirewallProfileInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		isDefault := false
		enabled := true
		ruleIDs := []string{uuid.New().String(), uuid.New().String()}
		input := FirewallProfileInput{
			Name:        "web-server",
			Description: "Profile for web servers",
			IsDefault:   &isDefault,
			Enabled:     &enabled,
			RuleIDs:     ruleIDs,
		}
		if input.Name != "web-server" {
			t.Error("Expected name to be 'web-server'")
		}
		if len(input.RuleIDs) != 2 {
			t.Error("Expected 2 rule IDs")
		}
	})
}

func TestFirewallFilter(t *testing.T) {
	t.Run("rule filter", func(t *testing.T) {
		chain := RuleChainInput
		enabled := true
		filter := &FirewallRuleFilter{
			Chain:   &chain,
			Enabled: &enabled,
		}
		if *filter.Chain != RuleChainInput {
			t.Error("Expected chain to be INPUT")
		}
		if !*filter.Enabled {
			t.Error("Expected enabled to be true")
		}
	})

	t.Run("profile filter", func(t *testing.T) {
		enabled := true
		filter := &FirewallProfileFilter{
			Enabled: &enabled,
		}
		if !*filter.Enabled {
			t.Error("Expected enabled to be true")
		}
	})
}
