package hypervisors

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

	var filter *HypervisorFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &HypervisorFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if status, ok := f["status"].(string); ok {
			s := HypervisorStatus(status)
			filter.Status = &s
		}
	}

	hypervisors, count, err := service.List(ctx, tenantID, filter, limit, offset)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"hypervisors":      hypervisors,
		"hypervisorsCount": count,
	}))
}

func handleGetHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	hypervisor, err := service.Get(ctx, tenantID, id)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"hypervisor": hypervisor,
	}))
}

func handleCreateHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input := &HypervisorInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if agentId, ok := inputRaw["agentId"].(string); ok {
		input.AgentID = agentId
	}
	if uri, ok := inputRaw["uri"].(string); ok {
		input.URI = uri
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}
	if input.AgentID == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("agentId is required"))
		return
	}
	if input.URI == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uri is required"))
		return
	}

	hypervisor, err := service.Create(ctx, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createHypervisor": hypervisor,
	}))
}

func handleUpdateHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input := &HypervisorInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if agentId, ok := inputRaw["agentId"].(string); ok {
		input.AgentID = agentId
	}
	if uri, ok := inputRaw["uri"].(string); ok {
		input.URI = uri
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}

	hypervisor, err := service.Update(ctx, tenantID, id, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"updateHypervisor": hypervisor,
	}))
}

func handleDeleteHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	// Get hypervisor info before deletion for audit
	hypervisor, _ := service.Get(ctx, tenantID, id)
	hypervisorName := ""
	if hypervisor != nil {
		hypervisorName = hypervisor.Name
	}

	if err := service.Delete(ctx, tenantID, id); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteHypervisor": true,
	}))
}

func handleTestHypervisorConnection(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	if err := service.TestConnection(ctx, token, tenantID, id, agentID); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"testHypervisorConnection": true,
	}))
}

func handleListLibvirtAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	client := csdcore.GetClient()
	agents, err := client.ListAgentsByCapability(ctx, token, "libvirt")
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"libvirtAgents": agents,
	}))
}

func handleListLibvirtDeployAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	// Filter by driver if provided
	driver, _ := variables["driver"].(string)

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
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"libvirtDeployAgents": enrichedAgents,
	}))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"libvirtDrivers": drivers,
	}))
}

func handleDeployHypervisor(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input := &DeployHypervisorInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if agentId, ok := inputRaw["agentId"].(string); ok {
		input.AgentID = agentId
	}
	if driver, ok := inputRaw["driver"].(string); ok {
		input.Driver = LibvirtDriver(driver)
	}

	// Validation
	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}
	if input.AgentID == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("agentId is required"))
		return
	}
	if input.Driver == "" {
		input.Driver = LibvirtDriverQEMU // Default to QEMU
	}

	hypervisor, err := service.Deploy(ctx, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deployHypervisor": hypervisor,
	}))
}

func handleBulkDeleteHypervisors(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idsRaw, ok := variables["ids"].([]interface{})
	if !ok || len(idsRaw) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("ids is required"))
		return
	}

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

	if len(ids) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("no valid ids provided"))
		return
	}

	deleted, err := service.BulkDelete(ctx, tenantID, ids)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"bulkDeleteHypervisors": deleted,
	}))
}
