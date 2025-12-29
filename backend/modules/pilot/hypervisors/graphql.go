package hypervisors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
)

func init() {
	service := NewService()

	// Queries
	graphql.RegisterQuery("hypervisors", "List all hypervisors", "csd-pilote.hypervisors.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListHypervisors(ctx, w, variables, service)
		})

	graphql.RegisterQuery("hypervisor", "Get a hypervisor by ID", "csd-pilote.hypervisors.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetHypervisor(ctx, w, variables, service)
		})

	graphql.RegisterQuery("domains", "List domains (VMs) on a hypervisor", "csd-pilote.domains.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListDomains(ctx, w, variables, service)
		})

	graphql.RegisterQuery("networks", "List networks on a hypervisor", "csd-pilote.networks.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListNetworks(ctx, w, variables, service)
		})

	graphql.RegisterQuery("storagePools", "List storage pools on a hypervisor", "csd-pilote.storage.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListStoragePools(ctx, w, variables, service)
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

	graphql.RegisterMutation("domainAction", "Perform action on a domain (start/stop/etc)", "csd-pilote.domains.power",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDomainAction(ctx, w, variables, service)
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
	if input.URI == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uri is required"))
		return
	}

	hypervisor, err := service.Create(ctx, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

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

	if err := service.Delete(ctx, tenantID, id); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

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

func handleListDomains(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorIDStr, ok := variables["hypervisorId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("hypervisorId is required"))
		return
	}

	hypervisorID, err := uuid.Parse(hypervisorIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid hypervisorId"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	domains, err := service.ListDomains(ctx, token, tenantID, hypervisorID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"domains": domains,
	}))
}

func handleListNetworks(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorIDStr, ok := variables["hypervisorId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("hypervisorId is required"))
		return
	}

	hypervisorID, err := uuid.Parse(hypervisorIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid hypervisorId"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	networks, err := service.ListNetworks(ctx, token, tenantID, hypervisorID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"networks": networks,
	}))
}

func handleListStoragePools(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorIDStr, ok := variables["hypervisorId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("hypervisorId is required"))
		return
	}

	hypervisorID, err := uuid.Parse(hypervisorIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid hypervisorId"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	pools, err := service.ListStoragePools(ctx, token, tenantID, hypervisorID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"storagePools": pools,
	}))
}

func handleDomainAction(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorIDStr, ok := variables["hypervisorId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("hypervisorId is required"))
		return
	}

	hypervisorID, err := uuid.Parse(hypervisorIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid hypervisorId"))
		return
	}

	domainName, ok := variables["domainName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("domainName is required"))
		return
	}

	action, ok := variables["action"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("action is required"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	if err := service.DomainAction(ctx, token, tenantID, hypervisorID, agentID, domainName, action); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"domainAction": true,
	}))
}
