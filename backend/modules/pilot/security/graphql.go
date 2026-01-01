package security

import (
	"context"
	"net/http"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
)

func init() {
	service := NewService()

	// ========================================
	// Firewall Rules Queries
	// ========================================

	graphql.RegisterQuery("securityRules", "List all firewall rules", "csd-pilote.security.rules.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListRules(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityRule", "Get a firewall rule by ID", "csd-pilote.security.rules.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetRule(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityRulesCount", "Count firewall rules", "csd-pilote.security.rules.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCountRules(ctx, w, variables, service)
		})

	// ========================================
	// Firewall Rules Mutations
	// ========================================

	graphql.RegisterMutation("createSecurityRule", "Create a new firewall rule", "csd-pilote.security.rules.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateRule(ctx, w, variables, service)
		})

	graphql.RegisterMutation("updateSecurityRule", "Update a firewall rule", "csd-pilote.security.rules.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleUpdateRule(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteSecurityRule", "Delete a firewall rule", "csd-pilote.security.rules.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteRule(ctx, w, variables, service)
		})

	graphql.RegisterMutation("bulkDeleteSecurityRules", "Delete multiple firewall rules", "csd-pilote.security.rules.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleBulkDeleteRules(ctx, w, variables, service)
		})

	// ========================================
	// Firewall Profiles Queries
	// ========================================

	graphql.RegisterQuery("securityProfiles", "List all firewall profiles", "csd-pilote.security.profiles.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListProfiles(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityProfile", "Get a firewall profile by ID", "csd-pilote.security.profiles.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetProfile(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityProfilesCount", "Count firewall profiles", "csd-pilote.security.profiles.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCountProfiles(ctx, w, variables, service)
		})

	// ========================================
	// Firewall Profiles Mutations
	// ========================================

	graphql.RegisterMutation("createSecurityProfile", "Create a new firewall profile", "csd-pilote.security.profiles.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateProfile(ctx, w, variables, service)
		})

	graphql.RegisterMutation("updateSecurityProfile", "Update a firewall profile", "csd-pilote.security.profiles.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleUpdateProfile(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteSecurityProfile", "Delete a firewall profile", "csd-pilote.security.profiles.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteProfile(ctx, w, variables, service)
		})

	graphql.RegisterMutation("addRulesToSecurityProfile", "Add rules to a profile", "csd-pilote.security.profiles.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleAddRulesToProfile(ctx, w, variables, service)
		})

	graphql.RegisterMutation("removeRulesFromSecurityProfile", "Remove rules from a profile", "csd-pilote.security.profiles.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleRemoveRulesFromProfile(ctx, w, variables, service)
		})

	// ========================================
	// Firewall Templates Queries
	// ========================================

	graphql.RegisterQuery("securityTemplates", "List all firewall templates", "csd-pilote.security.templates.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListTemplates(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityTemplate", "Get a firewall template by ID", "csd-pilote.security.templates.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetTemplate(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityTemplatesCount", "Count firewall templates", "csd-pilote.security.templates.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCountTemplates(ctx, w, variables, service)
		})

	// ========================================
	// Firewall Templates Mutations
	// ========================================

	graphql.RegisterMutation("createSecurityTemplate", "Create a new firewall template", "csd-pilote.security.templates.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateTemplate(ctx, w, variables, service)
		})

	graphql.RegisterMutation("updateSecurityTemplate", "Update a firewall template", "csd-pilote.security.templates.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleUpdateTemplate(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteSecurityTemplate", "Delete a firewall template", "csd-pilote.security.templates.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteTemplate(ctx, w, variables, service)
		})

	graphql.RegisterMutation("applySecurityTemplate", "Apply a template to a profile", "csd-pilote.security.templates.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleApplyTemplate(ctx, w, variables, service)
		})

	// ========================================
	// Firewall Deployments Queries
	// ========================================

	graphql.RegisterQuery("securityDeployments", "List all firewall deployments", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListDeployments(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityDeployment", "Get a firewall deployment by ID", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetDeployment(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityDeploymentsCount", "Count firewall deployments", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCountDeployments(ctx, w, variables, service)
		})

	graphql.RegisterQuery("securityAgents", "List agents that support nftables", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListSecurityAgents(ctx, w, variables, service)
		})

	// ========================================
	// Firewall Deployments Mutations
	// ========================================

	graphql.RegisterMutation("deploySecurityProfile", "Deploy a profile to an agent", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeployProfile(ctx, w, variables, service)
		})

	graphql.RegisterMutation("rollbackSecurityDeployment", "Rollback a deployment", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleRollbackDeployment(ctx, w, variables, service)
		})

	graphql.RegisterMutation("auditSecurityDeployment", "Audit firewall state on an agent", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleAuditDeployment(ctx, w, variables, service)
		})

	graphql.RegisterMutation("flushSecurityRules", "Flush all firewall rules on an agent", "csd-pilote.security.deploy",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleFlushRules(ctx, w, variables, service)
		})

	// ========================================
	// Import/Export Mutations
	// ========================================

	graphql.RegisterQuery("exportSecurityProfile", "Export a profile with its rules", "csd-pilote.security.profiles.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleExportProfile(ctx, w, variables, service)
		})

	graphql.RegisterMutation("importSecurityProfile", "Import a profile from JSON", "csd-pilote.security.profiles.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleImportProfile(ctx, w, variables, service)
		})
}

// ========================================
// Firewall Rules Handlers
// ========================================

func handleListRules(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	limit, offset := graphql.ParsePagination(variables)

	var filter *FirewallRuleFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallRuleFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if chain, ok := f["chain"].(string); ok {
			if err := graphql.ValidateEnum(chain, graphql.RuleChainValues, "chain"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			c := RuleChain(chain)
			filter.Chain = &c
		}
		if protocol, ok := f["protocol"].(string); ok {
			if err := graphql.ValidateEnum(protocol, graphql.RuleProtocolValues, "protocol"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			p := RuleProtocol(protocol)
			filter.Protocol = &p
		}
		if action, ok := f["action"].(string); ok {
			if err := graphql.ValidateEnum(action, graphql.RuleActionValues, "action"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			a := RuleAction(action)
			filter.Action = &a
		}
		if enabled, ok := f["enabled"].(bool); ok {
			filter.Enabled = &enabled
		}
	}

	rules, count, err := service.ListRules(ctx, tenantID, filter, limit, offset)
	if err != nil {
		graphql.WriteError(w, err, "list security rules")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityRules":      rules,
		"securityRulesCount": count,
	})
}

func handleGetRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	rule, err := service.GetRule(ctx, tenantID, id)
	if err != nil {
		graphql.WriteError(w, err, "get security rule")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityRule": rule,
	})
}

func handleCountRules(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	count, err := service.CountRules(ctx, tenantID)
	if err != nil {
		graphql.WriteError(w, err, "count security rules")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityRulesCount": count,
	})
}

func handleCreateRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseRuleInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate required field
	v := validation.NewValidator()
	v.Required("name", input.Name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	rule, err := service.CreateRule(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "create security rule")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"createSecurityRule": rule,
	})
}

func handleUpdateRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseRuleInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	rule, err := service.UpdateRule(ctx, token, tenantID, id, input)
	if err != nil {
		graphql.WriteError(w, err, "update security rule")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"updateSecurityRule": rule,
	})
}

func handleDeleteRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	if err := service.DeleteRule(ctx, token, tenantID, id); err != nil {
		graphql.WriteError(w, err, "delete security rule")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteSecurityRule": true,
	})
}

func handleBulkDeleteRules(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	ids, err := graphql.ParseBulkUUIDs(variables, "ids")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	deleted, err := service.BulkDeleteRules(ctx, tenantID, ids)
	if err != nil {
		graphql.WriteError(w, err, "bulk delete security rules")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"bulkDeleteSecurityRules": deleted,
	})
}

// ========================================
// Firewall Profiles Handlers
// ========================================

func handleListProfiles(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	limit, offset := graphql.ParsePagination(variables)

	var filter *FirewallProfileFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallProfileFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if isDefault, ok := f["isDefault"].(bool); ok {
			filter.IsDefault = &isDefault
		}
		if enabled, ok := f["enabled"].(bool); ok {
			filter.Enabled = &enabled
		}
	}

	profiles, count, err := service.ListProfiles(ctx, tenantID, filter, limit, offset)
	if err != nil {
		graphql.WriteError(w, err, "list security profiles")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityProfiles":      profiles,
		"securityProfilesCount": count,
	})
}

func handleGetProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	profile, err := service.GetProfileWithRules(ctx, tenantID, id)
	if err != nil {
		graphql.WriteError(w, err, "get security profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityProfile": profile,
	})
}

func handleCountProfiles(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	count, err := service.CountProfiles(ctx, tenantID)
	if err != nil {
		graphql.WriteError(w, err, "count security profiles")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityProfilesCount": count,
	})
}

func handleCreateProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseProfileInputWithValidation(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate required fields
	v := validation.NewValidator()
	v.Required("name", input.Name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	profile, err := service.CreateProfile(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "create security profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"createSecurityProfile": profile,
	})
}

func handleUpdateProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseProfileInputWithValidation(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	profile, err := service.UpdateProfile(ctx, token, tenantID, id, input)
	if err != nil {
		graphql.WriteError(w, err, "update security profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"updateSecurityProfile": profile,
	})
}

func handleDeleteProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	if err := service.DeleteProfile(ctx, token, tenantID, id); err != nil {
		graphql.WriteError(w, err, "delete security profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteSecurityProfile": true,
	})
}

func handleAddRulesToProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	profileID, err := graphql.ParseUUID(variables, "profileId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	ruleIDs, err := graphql.ParseBulkUUIDs(variables, "ruleIds")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	if err := service.AddRulesToProfile(ctx, tenantID, profileID, ruleIDs); err != nil {
		graphql.WriteError(w, err, "add rules to profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"addRulesToSecurityProfile": true,
	})
}

func handleRemoveRulesFromProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	profileID, err := graphql.ParseUUID(variables, "profileId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	ruleIDs, err := graphql.ParseBulkUUIDs(variables, "ruleIds")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	if err := service.RemoveRulesFromProfile(ctx, tenantID, profileID, ruleIDs); err != nil {
		graphql.WriteError(w, err, "remove rules from profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"removeRulesFromSecurityProfile": true,
	})
}

// ========================================
// Firewall Templates Handlers
// ========================================

func handleListTemplates(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	limit, offset := graphql.ParsePagination(variables)

	var filter *FirewallTemplateFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallTemplateFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if category, ok := f["category"].(string); ok {
			if err := graphql.ValidateEnum(category, graphql.TemplateCategoryValues, "category"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			c := TemplateCategory(category)
			filter.Category = &c
		}
		if isBuiltIn, ok := f["isBuiltIn"].(bool); ok {
			filter.IsBuiltIn = &isBuiltIn
		}
	}

	templates, count, err := service.ListTemplates(ctx, tenantID, filter, limit, offset)
	if err != nil {
		graphql.WriteError(w, err, "list security templates")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityTemplates":      templates,
		"securityTemplatesCount": count,
	})
}

func handleGetTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	template, err := service.GetTemplate(ctx, tenantID, id)
	if err != nil {
		graphql.WriteError(w, err, "get security template")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityTemplate": template,
	})
}

func handleCountTemplates(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	count, err := service.CountTemplates(ctx, tenantID)
	if err != nil {
		graphql.WriteError(w, err, "count security templates")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityTemplatesCount": count,
	})
}

func handleCreateTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseTemplateInputWithValidation(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate required fields
	v := validation.NewValidator()
	v.Required("name", input.Name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	template, err := service.CreateTemplate(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "create security template")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"createSecurityTemplate": template,
	})
}

func handleUpdateTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseTemplateInputWithValidation(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	template, err := service.UpdateTemplate(ctx, token, tenantID, id, input)
	if err != nil {
		graphql.WriteError(w, err, "update security template")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"updateSecurityTemplate": template,
	})
}

func handleDeleteTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	if err := service.DeleteTemplate(ctx, token, tenantID, id); err != nil {
		graphql.WriteError(w, err, "delete security template")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteSecurityTemplate": true,
	})
}

func handleApplyTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	templateID, err := graphql.ParseUUID(variables, "templateId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	profileID, err := graphql.ParseUUID(variables, "profileId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	if err := service.ApplyTemplateToProfile(ctx, token, tenantID, user.UserID, templateID, profileID); err != nil {
		graphql.WriteError(w, err, "apply security template")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"applySecurityTemplate": true,
	})
}

// ========================================
// Firewall Deployments Handlers
// ========================================

func handleListDeployments(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	limit, offset := graphql.ParsePagination(variables)

	var filter *FirewallDeploymentFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallDeploymentFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if profileId, ok := f["profileId"].(string); ok {
			v := validation.NewValidator()
			v.UUID("profileId", profileId)
			if v.HasErrors() {
				graphql.WriteValidationError(w, v.FirstError())
				return
			}
			filter.ProfileID = &profileId
		}
		if agentId, ok := f["agentId"].(string); ok {
			v := validation.NewValidator()
			v.UUID("agentId", agentId)
			if v.HasErrors() {
				graphql.WriteValidationError(w, v.FirstError())
				return
			}
			filter.AgentID = &agentId
		}
		if action, ok := f["action"].(string); ok {
			if err := graphql.ValidateEnum(action, graphql.DeploymentStatusValues, "action"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			a := DeploymentAction(action)
			filter.Action = &a
		}
		if status, ok := f["status"].(string); ok {
			if err := graphql.ValidateEnum(status, graphql.DeploymentStatusValues, "status"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			s := DeploymentStatus(status)
			filter.Status = &s
		}
	}

	deployments, count, err := service.ListDeployments(ctx, tenantID, filter, limit, offset)
	if err != nil {
		graphql.WriteError(w, err, "list security deployments")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityDeployments":      deployments,
		"securityDeploymentsCount": count,
	})
}

func handleGetDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	deployment, err := service.GetDeployment(ctx, tenantID, id)
	if err != nil {
		graphql.WriteError(w, err, "get security deployment")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityDeployment": deployment,
	})
}

func handleCountDeployments(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	count, err := service.CountDeployments(ctx, tenantID)
	if err != nil {
		graphql.WriteError(w, err, "count security deployments")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityDeploymentsCount": count,
	})
}

func handleListSecurityAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	client := csdcore.GetClient()
	agents, err := client.ListAgentsByCapability(ctx, token, "nftables")
	if err != nil {
		graphql.WriteError(w, err, "list security agents")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"securityAgents": agents,
	})
}

func handleDeployProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	profileIDStr, err := graphql.ParseStringRequired(variables, "profileId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}
	v := validation.NewValidator()
	v.UUID("profileId", profileIDStr)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	agentIDStr, err := graphql.ParseStringRequired(variables, "agentId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}
	v = validation.NewValidator()
	v.UUID("agentId", agentIDStr)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	// Check for dry-run mode
	dryRun := graphql.ParseBool(variables, "dryRun", false)

	input := &DeploymentInput{
		ProfileID: profileIDStr,
		AgentID:   agentIDStr,
		Action:    DeploymentActionApply,
		DryRun:    dryRun,
	}

	deployment, err := service.DeployProfile(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "deploy security profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deploySecurityProfile": deployment,
	})
}

func handleRollbackDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	deploymentID, err := graphql.ParseUUID(variables, "deploymentId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	rollback, err := service.RollbackDeployment(ctx, token, tenantID, user.UserID, deploymentID)
	if err != nil {
		graphql.WriteError(w, err, "rollback security deployment")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"rollbackSecurityDeployment": rollback,
	})
}

func handleAuditDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	agentID, err := graphql.ParseUUID(variables, "agentId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	audit, err := service.AuditDeployment(ctx, token, tenantID, user.UserID, agentID)
	if err != nil {
		graphql.WriteError(w, err, "audit security deployment")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"auditSecurityDeployment": audit,
	})
}

func handleFlushRules(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	agentID, err := graphql.ParseUUID(variables, "agentId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	flush, err := service.FlushRules(ctx, token, tenantID, user.UserID, agentID)
	if err != nil {
		graphql.WriteError(w, err, "flush security rules")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"flushSecurityRules": flush,
	})
}

// ========================================
// Import/Export Handlers
// ========================================

func handleExportProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	profileID, err := graphql.ParseUUID(variables, "profileId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	export, err := service.ExportProfile(ctx, token, tenantID, profileID)
	if err != nil {
		graphql.WriteError(w, err, "export security profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"exportSecurityProfile": export,
	})
}

func handleImportProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseProfileImportInputWithValidation(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate required fields
	v := validation.NewValidator()
	v.Required("name", input.Name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	profile, err := service.ImportProfile(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "import security profile")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"importSecurityProfile": profile,
	})
}

func parseProfileImportInputWithValidation(inputRaw map[string]interface{}) (*ProfileImportInput, error) {
	v := validation.NewValidator()
	input := &ProfileImportInput{}

	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		v.MaxLength("description", description, validation.MaxDescriptionLength)
		input.Description = description
	}

	if rules, ok := inputRaw["rules"].([]interface{}); ok {
		// Limit number of rules that can be imported
		v.MaxItems("rules", len(rules), validation.MaxBulkIDs)
		input.Rules = make([]TemplateRuleDefinition, 0, len(rules))
		for i, r := range rules {
			if ruleMap, ok := r.(map[string]interface{}); ok {
				rule := TemplateRuleDefinition{}
				if name, ok := ruleMap["name"].(string); ok {
					v.MaxLength("rules.name", name, validation.MaxNameLength).SafeString("rules.name", name)
					rule.Name = name
				}
				if description, ok := ruleMap["description"].(string); ok {
					v.MaxLength("rules.description", description, validation.MaxDescriptionLength)
					rule.Description = description
				}
				if chain, ok := ruleMap["chain"].(string); ok {
					if err := graphql.ValidateEnum(chain, graphql.RuleChainValues, "rules.chain"); err != nil {
						return nil, err
					}
					rule.Chain = RuleChain(chain)
				}
				if priority, ok := ruleMap["priority"].(float64); ok {
					p := int(priority)
					v.Range("rules.priority", p, 0, 65535)
					rule.Priority = p
				}
				if protocol, ok := ruleMap["protocol"].(string); ok {
					if err := graphql.ValidateEnum(protocol, graphql.RuleProtocolValues, "rules.protocol"); err != nil {
						return nil, err
					}
					rule.Protocol = RuleProtocol(protocol)
				}
				if sourceIp, ok := ruleMap["sourceIp"].(string); ok {
					if sourceIp != "" && sourceIp != "any" {
						v.SafeString("rules.sourceIp", sourceIp)
					}
					rule.SourceIP = sourceIp
				}
				if sourcePort, ok := ruleMap["sourcePort"].(string); ok {
					if sourcePort != "" {
						v.PortRange("rules.sourcePort", sourcePort)
					}
					rule.SourcePort = sourcePort
				}
				if destIp, ok := ruleMap["destIp"].(string); ok {
					if destIp != "" && destIp != "any" {
						v.SafeString("rules.destIp", destIp)
					}
					rule.DestIP = destIp
				}
				if destPort, ok := ruleMap["destPort"].(string); ok {
					if destPort != "" {
						v.PortRange("rules.destPort", destPort)
					}
					rule.DestPort = destPort
				}
				if action, ok := ruleMap["action"].(string); ok {
					if err := graphql.ValidateEnum(action, graphql.RuleActionValues, "rules.action"); err != nil {
						return nil, err
					}
					rule.Action = RuleAction(action)
				}
				if comment, ok := ruleMap["comment"].(string); ok {
					v.MaxLength("rules.comment", comment, 255).SafeString("rules.comment", comment)
					rule.Comment = comment
				}
				input.Rules = append(input.Rules, rule)
				// Stop early if we've hit too many errors
				if i > 0 && v.HasErrors() {
					break
				}
			}
		}
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}
	return input, nil
}

// ========================================
// Helper Functions
// ========================================

func parseRuleInput(inputRaw map[string]interface{}) (*FirewallRuleInput, error) {
	input := &FirewallRuleInput{}
	v := validation.NewValidator()

	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		v.MaxLength("description", description, validation.MaxDescriptionLength)
		input.Description = description
	}
	if chain, ok := inputRaw["chain"].(string); ok {
		if err := graphql.ValidateEnum(chain, graphql.RuleChainValues, "chain"); err != nil {
			return nil, err
		}
		input.Chain = RuleChain(chain)
	}
	if priority, ok := inputRaw["priority"].(float64); ok {
		p := int(priority)
		v.Range("priority", p, 0, 65535)
		input.Priority = p
	}
	if protocol, ok := inputRaw["protocol"].(string); ok {
		if err := graphql.ValidateEnum(protocol, graphql.RuleProtocolValues, "protocol"); err != nil {
			return nil, err
		}
		input.Protocol = RuleProtocol(protocol)
	}
	if sourceIp, ok := inputRaw["sourceIp"].(string); ok {
		if sourceIp != "" && sourceIp != "any" {
			v.SafeString("sourceIp", sourceIp)
		}
		input.SourceIP = sourceIp
	}
	if sourcePort, ok := inputRaw["sourcePort"].(string); ok {
		if sourcePort != "" {
			v.PortRange("sourcePort", sourcePort)
		}
		input.SourcePort = sourcePort
	}
	if destIp, ok := inputRaw["destIp"].(string); ok {
		if destIp != "" && destIp != "any" {
			v.SafeString("destIp", destIp)
		}
		input.DestIP = destIp
	}
	if destPort, ok := inputRaw["destPort"].(string); ok {
		if destPort != "" {
			v.PortRange("destPort", destPort)
		}
		input.DestPort = destPort
	}
	if action, ok := inputRaw["action"].(string); ok {
		if err := graphql.ValidateEnum(action, graphql.RuleActionValues, "action"); err != nil {
			return nil, err
		}
		input.Action = RuleAction(action)
	}
	// Interface matching
	if inInterface, ok := inputRaw["inInterface"].(string); ok {
		v.MaxLength("inInterface", inInterface, 64).SafeString("inInterface", inInterface)
		input.InInterface = inInterface
	}
	if outInterface, ok := inputRaw["outInterface"].(string); ok {
		v.MaxLength("outInterface", outInterface, 64).SafeString("outInterface", outInterface)
		input.OutInterface = outInterface
	}
	// Connection tracking
	if ctState, ok := inputRaw["ctState"].(string); ok {
		v.MaxLength("ctState", ctState, 128).SafeString("ctState", ctState)
		input.CTState = ctState
	}
	// Rate limiting
	if rateLimit, ok := inputRaw["rateLimit"].(string); ok {
		v.MaxLength("rateLimit", rateLimit, 64).SafeString("rateLimit", rateLimit)
		input.RateLimit = rateLimit
	}
	if rateBurst, ok := inputRaw["rateBurst"].(float64); ok {
		rb := int(rateBurst)
		v.Range("rateBurst", rb, 0, 65535)
		input.RateBurst = rb
	}
	if limitOver, ok := inputRaw["limitOver"].(string); ok {
		v.MaxLength("limitOver", limitOver, 64).SafeString("limitOver", limitOver)
		input.LimitOver = limitOver
	}
	// NAT options
	if natToAddr, ok := inputRaw["natToAddr"].(string); ok {
		v.MaxLength("natToAddr", natToAddr, 128).SafeString("natToAddr", natToAddr)
		input.NatToAddr = natToAddr
	}
	if natToPort, ok := inputRaw["natToPort"].(string); ok {
		if natToPort != "" {
			v.PortRange("natToPort", natToPort)
		}
		input.NatToPort = natToPort
	}
	// Logging options
	if logPrefix, ok := inputRaw["logPrefix"].(string); ok {
		v.MaxLength("logPrefix", logPrefix, 64).SafeString("logPrefix", logPrefix)
		input.LogPrefix = logPrefix
	}
	if logLevel, ok := inputRaw["logLevel"].(string); ok {
		v.MaxLength("logLevel", logLevel, 32).SafeString("logLevel", logLevel)
		input.LogLevel = logLevel
	}
	if ruleExpr, ok := inputRaw["ruleExpr"].(string); ok {
		// Validate nftables expression for safety
		v.NftablesExpression("ruleExpr", ruleExpr)
		v.MaxLength("ruleExpr", ruleExpr, 2000)
		input.RuleExpr = ruleExpr
	}
	if comment, ok := inputRaw["comment"].(string); ok {
		v.MaxLength("comment", comment, 255).SafeString("comment", comment)
		input.Comment = comment
	}
	if enabled, ok := inputRaw["enabled"].(bool); ok {
		input.Enabled = &enabled
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}

	return input, nil
}

func parseProfileInputWithValidation(inputRaw map[string]interface{}) (*FirewallProfileInput, error) {
	v := validation.NewValidator()
	input := &FirewallProfileInput{}

	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		v.MaxLength("description", description, validation.MaxDescriptionLength)
		input.Description = description
	}
	if isDefault, ok := inputRaw["isDefault"].(bool); ok {
		input.IsDefault = &isDefault
	}
	if enabled, ok := inputRaw["enabled"].(bool); ok {
		input.Enabled = &enabled
	}
	if ruleIds, ok := inputRaw["ruleIds"].([]interface{}); ok {
		v.MaxItems("ruleIds", len(ruleIds), validation.MaxBulkIDs)
		input.RuleIDs = make([]string, 0, len(ruleIds))
		for _, id := range ruleIds {
			if idStr, ok := id.(string); ok {
				v.UUID("ruleIds", idStr)
				input.RuleIDs = append(input.RuleIDs, idStr)
			}
		}
	}
	// Default policies - validate against allowed values
	policyValues := []string{"accept", "drop", "reject", ""}
	if inputPolicy, ok := inputRaw["inputPolicy"].(string); ok {
		v.Enum("inputPolicy", inputPolicy, policyValues)
		input.InputPolicy = inputPolicy
	}
	if outputPolicy, ok := inputRaw["outputPolicy"].(string); ok {
		v.Enum("outputPolicy", outputPolicy, policyValues)
		input.OutputPolicy = outputPolicy
	}
	if forwardPolicy, ok := inputRaw["forwardPolicy"].(string); ok {
		v.Enum("forwardPolicy", forwardPolicy, policyValues)
		input.ForwardPolicy = forwardPolicy
	}
	// Features
	if enableNat, ok := inputRaw["enableNat"].(bool); ok {
		input.EnableNAT = &enableNat
	}
	if enableConntrack, ok := inputRaw["enableConntrack"].(bool); ok {
		input.EnableConntrack = &enableConntrack
	}
	if allowLoopback, ok := inputRaw["allowLoopback"].(bool); ok {
		input.AllowLoopback = &allowLoopback
	}
	if allowEstablished, ok := inputRaw["allowEstablished"].(bool); ok {
		input.AllowEstablished = &allowEstablished
	}
	if allowIcmpPing, ok := inputRaw["allowIcmpPing"].(bool); ok {
		input.AllowICMPPing = &allowIcmpPing
	}
	if enableIpv6, ok := inputRaw["enableIpv6"].(bool); ok {
		input.EnableIPv6 = &enableIpv6
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}
	return input, nil
}

func parseTemplateInputWithValidation(inputRaw map[string]interface{}) (*FirewallTemplateInput, error) {
	v := validation.NewValidator()
	input := &FirewallTemplateInput{}

	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		v.MaxLength("description", description, validation.MaxDescriptionLength)
		input.Description = description
	}
	if category, ok := inputRaw["category"].(string); ok {
		if err := graphql.ValidateEnum(category, graphql.TemplateCategoryValues, "category"); err != nil {
			return nil, err
		}
		input.Category = TemplateCategory(category)
	}

	if rules, ok := inputRaw["rules"].([]interface{}); ok {
		// Limit number of rules in a template
		v.MaxItems("rules", len(rules), validation.MaxBulkIDs)
		input.Rules = make([]TemplateRuleDefinition, 0, len(rules))
		for i, r := range rules {
			if ruleMap, ok := r.(map[string]interface{}); ok {
				rule := TemplateRuleDefinition{}
				if name, ok := ruleMap["name"].(string); ok {
					v.MaxLength("rules.name", name, validation.MaxNameLength).SafeString("rules.name", name)
					rule.Name = name
				}
				if description, ok := ruleMap["description"].(string); ok {
					v.MaxLength("rules.description", description, validation.MaxDescriptionLength)
					rule.Description = description
				}
				if chain, ok := ruleMap["chain"].(string); ok {
					if err := graphql.ValidateEnum(chain, graphql.RuleChainValues, "rules.chain"); err != nil {
						return nil, err
					}
					rule.Chain = RuleChain(chain)
				}
				if priority, ok := ruleMap["priority"].(float64); ok {
					p := int(priority)
					v.Range("rules.priority", p, 0, 65535)
					rule.Priority = p
				}
				if protocol, ok := ruleMap["protocol"].(string); ok {
					if err := graphql.ValidateEnum(protocol, graphql.RuleProtocolValues, "rules.protocol"); err != nil {
						return nil, err
					}
					rule.Protocol = RuleProtocol(protocol)
				}
				if sourceIp, ok := ruleMap["sourceIp"].(string); ok {
					if sourceIp != "" && sourceIp != "any" {
						v.SafeString("rules.sourceIp", sourceIp)
					}
					rule.SourceIP = sourceIp
				}
				if sourcePort, ok := ruleMap["sourcePort"].(string); ok {
					if sourcePort != "" {
						v.PortRange("rules.sourcePort", sourcePort)
					}
					rule.SourcePort = sourcePort
				}
				if destIp, ok := ruleMap["destIp"].(string); ok {
					if destIp != "" && destIp != "any" {
						v.SafeString("rules.destIp", destIp)
					}
					rule.DestIP = destIp
				}
				if destPort, ok := ruleMap["destPort"].(string); ok {
					if destPort != "" {
						v.PortRange("rules.destPort", destPort)
					}
					rule.DestPort = destPort
				}
				if action, ok := ruleMap["action"].(string); ok {
					if err := graphql.ValidateEnum(action, graphql.RuleActionValues, "rules.action"); err != nil {
						return nil, err
					}
					rule.Action = RuleAction(action)
				}
				if comment, ok := ruleMap["comment"].(string); ok {
					v.MaxLength("rules.comment", comment, 255).SafeString("rules.comment", comment)
					rule.Comment = comment
				}
				input.Rules = append(input.Rules, rule)
				// Stop early if we've hit too many errors
				if i > 0 && v.HasErrors() {
					break
				}
			}
		}
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}
	return input, nil
}
