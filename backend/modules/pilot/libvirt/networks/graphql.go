package networks

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
	graphql.RegisterQuery("libvirtNetworks", "List libvirt networks", "csd-pilote.networks.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListNetworks(ctx, w, variables, service)
		})

	graphql.RegisterQuery("libvirtNetwork", "Get a libvirt network", "csd-pilote.networks.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetNetwork(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("startLibvirtNetwork", "Start a network", "csd-pilote.networks.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleStartNetwork(ctx, w, variables, service)
		})

	graphql.RegisterMutation("stopLibvirtNetwork", "Stop a network", "csd-pilote.networks.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleStopNetwork(ctx, w, variables, service)
		})

	graphql.RegisterMutation("setLibvirtNetworkAutostart", "Set network autostart", "csd-pilote.networks.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleSetNetworkAutostart(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteLibvirtNetwork", "Delete a network", "csd-pilote.networks.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteNetwork(ctx, w, variables, service)
		})
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

	var filter *NetworkFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &NetworkFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if active, ok := f["active"].(bool); ok {
			filter.Active = &active
		}
	}

	networks, err := service.List(ctx, token, tenantID, hypervisorID, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"libvirtNetworks":      networks,
		"libvirtNetworksCount": len(networks),
	}))
}

func handleGetNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	network, err := service.Get(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"libvirtNetwork": network,
	}))
}

func handleStartNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	network, err := service.Start(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"startLibvirtNetwork": network,
	}))
}

func handleStopNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	network, err := service.Stop(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"stopLibvirtNetwork": network,
	}))
}

func handleSetNetworkAutostart(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	autostart, ok := variables["autostart"].(bool)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("autostart is required"))
		return
	}

	network, err := service.SetAutostart(ctx, token, tenantID, hypervisorID, name, autostart)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"setLibvirtNetworkAutostart": network,
	}))
}

func handleDeleteNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	if err := service.Delete(ctx, token, tenantID, hypervisorID, name); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteLibvirtNetwork": true,
	}))
}
