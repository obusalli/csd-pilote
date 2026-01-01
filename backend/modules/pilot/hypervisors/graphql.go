package hypervisors

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

	// Queries
	graphql.RegisterQuery("libvirtAgents", "List agents that support Libvirt management", "csd-pilote.hypervisors.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListLibvirtAgents(ctx, w, variables, service)
		})

	graphql.RegisterQuery("libvirtDeployAgents", "List agents where Libvirt can be deployed", "csd-pilote.hypervisors.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListLibvirtDeployAgents(ctx, w, variables, service)
		})

	graphql.RegisterQuery("libvirtDrivers", "List supported Libvirt drivers", "csd-pilote.hypervisors.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListLibvirtDrivers(ctx, w, variables, service)
		})

	graphql.RegisterQuery("hypervisors", "List all hypervisors", "csd-pilote.hypervisors.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListHypervisors(ctx, w, variables, service)
		})

	graphql.RegisterQuery("hypervisor", "Get a hypervisor by ID", "csd-pilote.hypervisors.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetHypervisor(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("createHypervisor", "Create a new hypervisor", "csd-pilote.hypervisors.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateHypervisor(ctx, w, variables, service)
		})

	graphql.RegisterMutation("updateHypervisor", "Update a hypervisor", "csd-pilote.hypervisors.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleUpdateHypervisor(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteHypervisor", "Delete a hypervisor", "csd-pilote.hypervisors.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteHypervisor(ctx, w, variables, service)
		})

	graphql.RegisterMutation("testHypervisorConnection", "Test hypervisor connection", "csd-pilote.hypervisors.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleTestHypervisorConnection(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deployHypervisor", "Deploy Libvirt on an agent", "csd-pilote.hypervisors.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeployHypervisor(ctx, w, variables, service)
		})

	graphql.RegisterMutation("bulkDeleteHypervisors", "Delete multiple hypervisors", "csd-pilote.hypervisors.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleBulkDeleteHypervisors(ctx, w, variables, service)
		})
}

func handleListHypervisors(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	limit, offset := graphql.ParsePagination(variables)

	var filter *HypervisorFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &HypervisorFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if status, ok := f["status"].(string); ok {
			if err := graphql.ValidateEnum(status, graphql.HypervisorStatusValues, "status"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			s := HypervisorStatus(status)
			filter.Status = &s
		}
		if mode, ok := f["mode"].(string); ok {
			if err := graphql.ValidateEnum(mode, graphql.HypervisorModeValues, "mode"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			m := HypervisorMode(mode)
			filter.Mode = &m
		}
	}

	hypervisors, count, err := service.List(ctx, tenantID, filter, limit, offset)
	if err != nil {
		graphql.WriteError(w, err, "list hypervisors")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"hypervisors":      hypervisors,
		"hypervisorsCount": count,
	})
}

func handleGetHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	hypervisor, err := service.Get(ctx, tenantID, id)
	if err != nil {
		graphql.WriteError(w, err, "get hypervisor")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"hypervisor": hypervisor,
	})
}

func handleCreateHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input, err := parseHypervisorInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate required fields
	v := validation.NewValidator()
	v.Required("name", input.Name).Required("agentId", input.AgentID).Required("uri", input.URI)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	hypervisor, err := service.Create(ctx, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "create hypervisor")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "CREATE_HYPERVISOR",
		ResourceType: "hypervisor",
		ResourceID:   hypervisor.ID.String(),
		Details: map[string]interface{}{
			"name":   hypervisor.Name,
			"mode":   hypervisor.Mode,
			"driver": hypervisor.Driver,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"createHypervisor": hypervisor,
	})
}

func handleUpdateHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input, err := parseHypervisorInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	hypervisor, err := service.Update(ctx, tenantID, id, input)
	if err != nil {
		graphql.WriteError(w, err, "update hypervisor")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "UPDATE_HYPERVISOR",
		ResourceType: "hypervisor",
		ResourceID:   hypervisor.ID.String(),
		Details: map[string]interface{}{
			"name": hypervisor.Name,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"updateHypervisor": hypervisor,
	})
}

func handleDeleteHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	// Get hypervisor info before deletion for audit
	hypervisor, _ := service.Get(ctx, tenantID, id)
	hypervisorName := ""
	if hypervisor != nil {
		hypervisorName = hypervisor.Name
	}

	if err := service.Delete(ctx, tenantID, id); err != nil {
		graphql.WriteError(w, err, "delete hypervisor")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DELETE_HYPERVISOR",
		ResourceType: "hypervisor",
		ResourceID:   id.String(),
		Details: map[string]interface{}{
			"name": hypervisorName,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteHypervisor": true,
	})
}

func handleTestHypervisorConnection(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	// agentId is optional - parse but don't fail if invalid
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	if err := service.TestConnection(ctx, token, tenantID, id, agentID); err != nil {
		graphql.WriteError(w, err, "test hypervisor connection")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"testHypervisorConnection": true,
	})
}

func handleListLibvirtAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	client := csdcore.GetClient()
	agents, err := client.ListAgentsByCapability(ctx, token, "libvirt")
	if err != nil {
		graphql.WriteError(w, err, "list libvirt agents")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"libvirtAgents": agents,
	})
}

func handleListLibvirtDeployAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	// Filter by driver if provided
	driver := graphql.ParseString(variables, "driver")
	if driver != "" {
		if err := graphql.ValidateEnum(driver, graphql.LibvirtDriverValues, "driver"); err != nil {
			graphql.WriteValidationError(w, err.Error())
			return
		}
	}

	client := csdcore.GetClient()
	var agents []csdcore.Agent
	var err error

	if driver != "" {
		// Filter by specific driver capability (e.g., "libvirt-deploy-qemu")
		capability := "libvirt-deploy-" + driver
		agents, err = client.ListAgentsByCapability(ctx, token, capability)
	} else {
		// Get all agents that can deploy any Libvirt driver
		agents, err = client.ListAgentsByCapabilityPrefix(ctx, token, "libvirt-deploy-")
	}

	if err != nil {
		graphql.WriteError(w, err, "list libvirt deploy agents")
		return
	}

	// Enrich agents with their supported drivers
	type AgentWithDrivers struct {
		csdcore.Agent
		SupportedDrivers []string `json:"supportedDrivers"`
	}

	enrichedAgents := make([]AgentWithDrivers, 0, len(agents))
	for _, agent := range agents {
		driverCaps := agent.GetCapabilitiesByPrefix("libvirt-deploy-")
		drivers := make([]string, 0, len(driverCaps))
		for _, d := range driverCaps {
			if len(d) > len("libvirt-deploy-") {
				drivers = append(drivers, d[len("libvirt-deploy-"):])
			}
		}
		enrichedAgents = append(enrichedAgents, AgentWithDrivers{
			Agent:            agent,
			SupportedDrivers: drivers,
		})
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"libvirtDeployAgents": enrichedAgents,
	})
}

func handleListLibvirtDrivers(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	drivers := []map[string]interface{}{
		{
			"id":          "qemu",
			"name":        "QEMU/KVM",
			"description": "Full virtualization with KVM acceleration",
			"capability":  "libvirt-deploy-qemu",
		},
		{
			"id":          "xen",
			"name":        "Xen",
			"description": "Xen hypervisor for paravirtualization",
			"capability":  "libvirt-deploy-xen",
		},
		{
			"id":          "lxc",
			"name":        "LXC",
			"description": "Linux Containers - OS-level virtualization",
			"capability":  "libvirt-deploy-lxc",
		},
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"libvirtDrivers": drivers,
	})
}

func handleDeployHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input, err := parseDeployHypervisorInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate required fields
	v := validation.NewValidator()
	v.Required("name", input.Name).Required("agentId", input.AgentID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	// Default driver
	if input.Driver == "" {
		input.Driver = LibvirtDriverQEMU
	}

	hypervisor, err := service.Deploy(ctx, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "deploy hypervisor")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DEPLOY_HYPERVISOR",
		ResourceType: "hypervisor",
		ResourceID:   hypervisor.ID.String(),
		Details: map[string]interface{}{
			"name":    hypervisor.Name,
			"driver":  hypervisor.Driver,
			"agentId": input.AgentID,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deployHypervisor": hypervisor,
	})
}

func handleBulkDeleteHypervisors(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	ids, err := graphql.ParseBulkUUIDs(variables, "ids")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	deleted, err := service.BulkDelete(ctx, tenantID, ids)
	if err != nil {
		graphql.WriteError(w, err, "bulk delete hypervisors")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "BULK_DELETE_HYPERVISORS",
		ResourceType: "hypervisor",
		ResourceID:   "",
		Details: map[string]interface{}{
			"count": deleted,
			"ids":   ids,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"bulkDeleteHypervisors": deleted,
	})
}

// ========================================
// Helper Functions
// ========================================

func parseHypervisorInput(inputRaw map[string]interface{}) (*HypervisorInput, error) {
	input := &HypervisorInput{}
	v := validation.NewValidator()

	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		v.MaxLength("description", description, validation.MaxDescriptionLength)
		input.Description = description
	}
	if agentId, ok := inputRaw["agentId"].(string); ok {
		v.UUID("agentId", agentId)
		input.AgentID = agentId
	}
	if uri, ok := inputRaw["uri"].(string); ok {
		v.MaxLength("uri", uri, 1024).SafeString("uri", uri)
		input.URI = uri
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		v.MaxLength("artifactKey", artifactKey, 255).SafeString("artifactKey", artifactKey)
		input.ArtifactKey = artifactKey
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}
	return input, nil
}

func parseDeployHypervisorInput(inputRaw map[string]interface{}) (*DeployHypervisorInput, error) {
	input := &DeployHypervisorInput{}
	v := validation.NewValidator()

	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		v.MaxLength("description", description, validation.MaxDescriptionLength)
		input.Description = description
	}
	if agentId, ok := inputRaw["agentId"].(string); ok {
		v.UUID("agentId", agentId)
		input.AgentID = agentId
	}
	if driver, ok := inputRaw["driver"].(string); ok {
		if err := graphql.ValidateEnum(driver, graphql.LibvirtDriverValues, "driver"); err != nil {
			return nil, err
		}
		input.Driver = LibvirtDriver(driver)
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}
	return input, nil
}
