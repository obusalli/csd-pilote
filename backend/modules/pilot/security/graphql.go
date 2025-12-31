package security

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
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
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	limit := 20
	offset := 0
	if l, ok := variables["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := variables["offset"].(float64); ok {
		offset = int(o)
	}

	var filter *FirewallRuleFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallRuleFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if chain, ok := f["chain"].(string); ok {
			c := RuleChain(chain)
			filter.Chain = &c
		}
		if protocol, ok := f["protocol"].(string); ok {
			p := RuleProtocol(protocol)
			filter.Protocol = &p
		}
		if action, ok := f["action"].(string); ok {
			a := RuleAction(action)
			filter.Action = &a
		}
		if enabled, ok := f["enabled"].(bool); ok {
			filter.Enabled = &enabled
		}
	}

	rules, count, err := service.ListRules(ctx, tenantID, filter, limit, offset)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityRules":      rules,
		"securityRulesCount": count,
	}))
}

func handleGetRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	rule, err := service.GetRule(ctx, tenantID, id)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityRule": rule,
	}))
}

func handleCountRules(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	count, err := service.CountRules(ctx, tenantID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityRulesCount": count,
	}))
}

func handleCreateRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := parseRuleInput(inputRaw)

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	rule, err := service.CreateRule(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createSecurityRule": rule,
	}))
}

func handleUpdateRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := parseRuleInput(inputRaw)

	rule, err := service.UpdateRule(ctx, token, tenantID, id, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"updateSecurityRule": rule,
	}))
}

func handleDeleteRule(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	if err := service.DeleteRule(ctx, token, tenantID, id); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteSecurityRule": true,
	}))
}

func handleBulkDeleteRules(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idsRaw, ok := variables["ids"].([]interface{})
	if !ok || len(idsRaw) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("ids is required"))
		return
	}

	ids := parseUUIDs(idsRaw)
	if len(ids) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("no valid ids provided"))
		return
	}

	deleted, err := service.BulkDeleteRules(ctx, tenantID, ids)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"bulkDeleteSecurityRules": deleted,
	}))
}

// ========================================
// Firewall Profiles Handlers
// ========================================

func handleListProfiles(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	limit := 20
	offset := 0
	if l, ok := variables["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := variables["offset"].(float64); ok {
		offset = int(o)
	}

	var filter *FirewallProfileFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallProfileFilter{}
		if search, ok := f["search"].(string); ok {
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
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityProfiles":      profiles,
		"securityProfilesCount": count,
	}))
}

func handleGetProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	profile, err := service.GetProfileWithRules(ctx, tenantID, id)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityProfile": profile,
	}))
}

func handleCountProfiles(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	count, err := service.CountProfiles(ctx, tenantID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityProfilesCount": count,
	}))
}

func handleCreateProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := parseProfileInput(inputRaw)

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	profile, err := service.CreateProfile(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createSecurityProfile": profile,
	}))
}

func handleUpdateProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := parseProfileInput(inputRaw)

	profile, err := service.UpdateProfile(ctx, token, tenantID, id, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"updateSecurityProfile": profile,
	}))
}

func handleDeleteProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	if err := service.DeleteProfile(ctx, token, tenantID, id); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteSecurityProfile": true,
	}))
}

func handleAddRulesToProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	profileIDStr, ok := variables["profileId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("profileId is required"))
		return
	}

	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid profileId"))
		return
	}

	ruleIDsRaw, ok := variables["ruleIds"].([]interface{})
	if !ok || len(ruleIDsRaw) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("ruleIds is required"))
		return
	}

	ruleIDs := parseUUIDs(ruleIDsRaw)

	if err := service.AddRulesToProfile(ctx, tenantID, profileID, ruleIDs); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"addRulesToSecurityProfile": true,
	}))
}

func handleRemoveRulesFromProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	profileIDStr, ok := variables["profileId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("profileId is required"))
		return
	}

	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid profileId"))
		return
	}

	ruleIDsRaw, ok := variables["ruleIds"].([]interface{})
	if !ok || len(ruleIDsRaw) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("ruleIds is required"))
		return
	}

	ruleIDs := parseUUIDs(ruleIDsRaw)

	if err := service.RemoveRulesFromProfile(ctx, tenantID, profileID, ruleIDs); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"removeRulesFromSecurityProfile": true,
	}))
}

// ========================================
// Firewall Templates Handlers
// ========================================

func handleListTemplates(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	limit := 20
	offset := 0
	if l, ok := variables["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := variables["offset"].(float64); ok {
		offset = int(o)
	}

	var filter *FirewallTemplateFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallTemplateFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if category, ok := f["category"].(string); ok {
			c := TemplateCategory(category)
			filter.Category = &c
		}
		if isBuiltIn, ok := f["isBuiltIn"].(bool); ok {
			filter.IsBuiltIn = &isBuiltIn
		}
	}

	templates, count, err := service.ListTemplates(ctx, tenantID, filter, limit, offset)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityTemplates":      templates,
		"securityTemplatesCount": count,
	}))
}

func handleGetTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	template, err := service.GetTemplate(ctx, tenantID, id)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityTemplate": template,
	}))
}

func handleCountTemplates(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	count, err := service.CountTemplates(ctx, tenantID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityTemplatesCount": count,
	}))
}

func handleCreateTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := parseTemplateInput(inputRaw)

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	template, err := service.CreateTemplate(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createSecurityTemplate": template,
	}))
}

func handleUpdateTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := parseTemplateInput(inputRaw)

	template, err := service.UpdateTemplate(ctx, token, tenantID, id, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"updateSecurityTemplate": template,
	}))
}

func handleDeleteTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	if err := service.DeleteTemplate(ctx, token, tenantID, id); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteSecurityTemplate": true,
	}))
}

func handleApplyTemplate(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	templateIDStr, ok := variables["templateId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("templateId is required"))
		return
	}

	profileIDStr, ok := variables["profileId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("profileId is required"))
		return
	}

	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid templateId"))
		return
	}

	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid profileId"))
		return
	}

	if err := service.ApplyTemplateToProfile(ctx, token, tenantID, user.UserID, templateID, profileID); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"applySecurityTemplate": true,
	}))
}

// ========================================
// Firewall Deployments Handlers
// ========================================

func handleListDeployments(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	limit := 20
	offset := 0
	if l, ok := variables["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := variables["offset"].(float64); ok {
		offset = int(o)
	}

	var filter *FirewallDeploymentFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &FirewallDeploymentFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if profileId, ok := f["profileId"].(string); ok {
			filter.ProfileID = &profileId
		}
		if agentId, ok := f["agentId"].(string); ok {
			filter.AgentID = &agentId
		}
		if action, ok := f["action"].(string); ok {
			a := DeploymentAction(action)
			filter.Action = &a
		}
		if status, ok := f["status"].(string); ok {
			s := DeploymentStatus(status)
			filter.Status = &s
		}
	}

	deployments, count, err := service.ListDeployments(ctx, tenantID, filter, limit, offset)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityDeployments":      deployments,
		"securityDeploymentsCount": count,
	}))
}

func handleGetDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	deployment, err := service.GetDeployment(ctx, tenantID, id)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityDeployment": deployment,
	}))
}

func handleCountDeployments(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	count, err := service.CountDeployments(ctx, tenantID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityDeploymentsCount": count,
	}))
}

func handleListSecurityAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	client := csdcore.GetClient()
	agents, err := client.ListAgentsByCapability(ctx, token, "nftables")
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"securityAgents": agents,
	}))
}

func handleDeployProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	profileIDStr, ok := variables["profileId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("profileId is required"))
		return
	}

	agentIDStr, ok := variables["agentId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("agentId is required"))
		return
	}

	// Check for dry-run mode
	dryRun := false
	if dr, ok := variables["dryRun"].(bool); ok {
		dryRun = dr
	}

	input := &DeploymentInput{
		ProfileID: profileIDStr,
		AgentID:   agentIDStr,
		Action:    DeploymentActionApply,
		DryRun:    dryRun,
	}

	deployment, err := service.DeployProfile(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deploySecurityProfile": deployment,
	}))
}

func handleRollbackDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	deploymentIDStr, ok := variables["deploymentId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("deploymentId is required"))
		return
	}

	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid deploymentId"))
		return
	}

	rollback, err := service.RollbackDeployment(ctx, token, tenantID, user.UserID, deploymentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"rollbackSecurityDeployment": rollback,
	}))
}

func handleAuditDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	agentIDStr, ok := variables["agentId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("agentId is required"))
		return
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid agentId"))
		return
	}

	audit, err := service.AuditDeployment(ctx, token, tenantID, user.UserID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"auditSecurityDeployment": audit,
	}))
}

func handleFlushRules(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	agentIDStr, ok := variables["agentId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("agentId is required"))
		return
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid agentId"))
		return
	}

	flush, err := service.FlushRules(ctx, token, tenantID, user.UserID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"flushSecurityRules": flush,
	}))
}

// ========================================
// Import/Export Handlers
// ========================================

func handleExportProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	profileIDStr, ok := variables["profileId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("profileId is required"))
		return
	}

	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid profileId"))
		return
	}

	export, err := service.ExportProfile(ctx, token, tenantID, profileID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"exportSecurityProfile": export,
	}))
}

func handleImportProfile(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := parseProfileImportInput(inputRaw)

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	profile, err := service.ImportProfile(ctx, token, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"importSecurityProfile": profile,
	}))
}

func parseProfileImportInput(inputRaw map[string]interface{}) *ProfileImportInput {
	input := &ProfileImportInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if rules, ok := inputRaw["rules"].([]interface{}); ok {
		input.Rules = make([]TemplateRuleDefinition, 0, len(rules))
		for _, r := range rules {
			if ruleMap, ok := r.(map[string]interface{}); ok {
				rule := TemplateRuleDefinition{}
				if name, ok := ruleMap["name"].(string); ok {
					rule.Name = name
				}
				if description, ok := ruleMap["description"].(string); ok {
					rule.Description = description
				}
				if chain, ok := ruleMap["chain"].(string); ok {
					rule.Chain = RuleChain(chain)
				}
				if priority, ok := ruleMap["priority"].(float64); ok {
					rule.Priority = int(priority)
				}
				if protocol, ok := ruleMap["protocol"].(string); ok {
					rule.Protocol = RuleProtocol(protocol)
				}
				if sourceIp, ok := ruleMap["sourceIp"].(string); ok {
					rule.SourceIP = sourceIp
				}
				if sourcePort, ok := ruleMap["sourcePort"].(string); ok {
					rule.SourcePort = sourcePort
				}
				if destIp, ok := ruleMap["destIp"].(string); ok {
					rule.DestIP = destIp
				}
				if destPort, ok := ruleMap["destPort"].(string); ok {
					rule.DestPort = destPort
				}
				if action, ok := ruleMap["action"].(string); ok {
					rule.Action = RuleAction(action)
				}
				if comment, ok := ruleMap["comment"].(string); ok {
					rule.Comment = comment
				}
				input.Rules = append(input.Rules, rule)
			}
		}
	}
	return input
}

// ========================================
// Helper Functions
// ========================================

func parseRuleInput(inputRaw map[string]interface{}) *FirewallRuleInput {
	input := &FirewallRuleInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if chain, ok := inputRaw["chain"].(string); ok {
		input.Chain = RuleChain(chain)
	}
	if priority, ok := inputRaw["priority"].(float64); ok {
		input.Priority = int(priority)
	}
	if protocol, ok := inputRaw["protocol"].(string); ok {
		input.Protocol = RuleProtocol(protocol)
	}
	if sourceIp, ok := inputRaw["sourceIp"].(string); ok {
		input.SourceIP = sourceIp
	}
	if sourcePort, ok := inputRaw["sourcePort"].(string); ok {
		input.SourcePort = sourcePort
	}
	if destIp, ok := inputRaw["destIp"].(string); ok {
		input.DestIP = destIp
	}
	if destPort, ok := inputRaw["destPort"].(string); ok {
		input.DestPort = destPort
	}
	if action, ok := inputRaw["action"].(string); ok {
		input.Action = RuleAction(action)
	}
	// Interface matching
	if inInterface, ok := inputRaw["inInterface"].(string); ok {
		input.InInterface = inInterface
	}
	if outInterface, ok := inputRaw["outInterface"].(string); ok {
		input.OutInterface = outInterface
	}
	// Connection tracking
	if ctState, ok := inputRaw["ctState"].(string); ok {
		input.CTState = ctState
	}
	// Rate limiting
	if rateLimit, ok := inputRaw["rateLimit"].(string); ok {
		input.RateLimit = rateLimit
	}
	if rateBurst, ok := inputRaw["rateBurst"].(float64); ok {
		input.RateBurst = int(rateBurst)
	}
	if limitOver, ok := inputRaw["limitOver"].(string); ok {
		input.LimitOver = limitOver
	}
	// NAT options
	if natToAddr, ok := inputRaw["natToAddr"].(string); ok {
		input.NatToAddr = natToAddr
	}
	if natToPort, ok := inputRaw["natToPort"].(string); ok {
		input.NatToPort = natToPort
	}
	// Logging options
	if logPrefix, ok := inputRaw["logPrefix"].(string); ok {
		input.LogPrefix = logPrefix
	}
	if logLevel, ok := inputRaw["logLevel"].(string); ok {
		input.LogLevel = logLevel
	}
	if ruleExpr, ok := inputRaw["ruleExpr"].(string); ok {
		input.RuleExpr = ruleExpr
	}
	if comment, ok := inputRaw["comment"].(string); ok {
		input.Comment = comment
	}
	if enabled, ok := inputRaw["enabled"].(bool); ok {
		input.Enabled = &enabled
	}
	return input
}

func parseProfileInput(inputRaw map[string]interface{}) *FirewallProfileInput {
	input := &FirewallProfileInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if isDefault, ok := inputRaw["isDefault"].(bool); ok {
		input.IsDefault = &isDefault
	}
	if enabled, ok := inputRaw["enabled"].(bool); ok {
		input.Enabled = &enabled
	}
	if ruleIds, ok := inputRaw["ruleIds"].([]interface{}); ok {
		input.RuleIDs = make([]string, 0, len(ruleIds))
		for _, id := range ruleIds {
			if idStr, ok := id.(string); ok {
				input.RuleIDs = append(input.RuleIDs, idStr)
			}
		}
	}
	// Default policies
	if inputPolicy, ok := inputRaw["inputPolicy"].(string); ok {
		input.InputPolicy = inputPolicy
	}
	if outputPolicy, ok := inputRaw["outputPolicy"].(string); ok {
		input.OutputPolicy = outputPolicy
	}
	if forwardPolicy, ok := inputRaw["forwardPolicy"].(string); ok {
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
	return input
}

func parseTemplateInput(inputRaw map[string]interface{}) *FirewallTemplateInput {
	input := &FirewallTemplateInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if category, ok := inputRaw["category"].(string); ok {
		input.Category = TemplateCategory(category)
	}
	if rules, ok := inputRaw["rules"].([]interface{}); ok {
		input.Rules = make([]TemplateRuleDefinition, 0, len(rules))
		for _, r := range rules {
			if ruleMap, ok := r.(map[string]interface{}); ok {
				rule := TemplateRuleDefinition{}
				if name, ok := ruleMap["name"].(string); ok {
					rule.Name = name
				}
				if description, ok := ruleMap["description"].(string); ok {
					rule.Description = description
				}
				if chain, ok := ruleMap["chain"].(string); ok {
					rule.Chain = RuleChain(chain)
				}
				if priority, ok := ruleMap["priority"].(float64); ok {
					rule.Priority = int(priority)
				}
				if protocol, ok := ruleMap["protocol"].(string); ok {
					rule.Protocol = RuleProtocol(protocol)
				}
				if sourceIp, ok := ruleMap["sourceIp"].(string); ok {
					rule.SourceIP = sourceIp
				}
				if sourcePort, ok := ruleMap["sourcePort"].(string); ok {
					rule.SourcePort = sourcePort
				}
				if destIp, ok := ruleMap["destIp"].(string); ok {
					rule.DestIP = destIp
				}
				if destPort, ok := ruleMap["destPort"].(string); ok {
					rule.DestPort = destPort
				}
				if action, ok := ruleMap["action"].(string); ok {
					rule.Action = RuleAction(action)
				}
				if comment, ok := ruleMap["comment"].(string); ok {
					rule.Comment = comment
				}
				input.Rules = append(input.Rules, rule)
			}
		}
	}
	return input
}

func parseUUIDs(idsRaw []interface{}) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(idsRaw))
	for _, idRaw := range idsRaw {
		idStr, ok := idRaw.(string)
		if !ok {
			continue
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}
