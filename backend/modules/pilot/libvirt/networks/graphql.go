package networks

import (
	"context"
	"net/http"

	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
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
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorID, err := graphql.ParseUUID(variables, "hypervisorId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	var filter *NetworkFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &NetworkFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if active, ok := f["active"].(bool); ok {
			filter.Active = &active
		}
	}

	networks, err := service.List(ctx, token, tenantID, hypervisorID, filter)
	if err != nil {
		graphql.WriteError(w, err, "list libvirt networks")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"libvirtNetworks":      networks,
		"libvirtNetworksCount": len(networks),
	})
}

func handleGetNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorID, err := graphql.ParseUUID(variables, "hypervisorId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate network name
	v := validation.NewValidator()
	v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	network, err := service.Get(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		graphql.WriteError(w, err, "get libvirt network")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"libvirtNetwork": network,
	})
}

func handleStartNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorID, err := graphql.ParseUUID(variables, "hypervisorId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	network, err := service.Start(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		graphql.WriteError(w, err, "start libvirt network")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"startLibvirtNetwork": network,
	})
}

func handleStopNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorID, err := graphql.ParseUUID(variables, "hypervisorId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	network, err := service.Stop(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		graphql.WriteError(w, err, "stop libvirt network")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"stopLibvirtNetwork": network,
	})
}

func handleSetNetworkAutostart(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorID, err := graphql.ParseUUID(variables, "hypervisorId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	autostart, ok := variables["autostart"].(bool)
	if !ok {
		graphql.WriteValidationError(w, "autostart is required")
		return
	}

	network, err := service.SetAutostart(ctx, token, tenantID, hypervisorID, name, autostart)
	if err != nil {
		graphql.WriteError(w, err, "set libvirt network autostart")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"setLibvirtNetworkAutostart": network,
	})
}

func handleDeleteNetwork(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorID, err := graphql.ParseUUID(variables, "hypervisorId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	if err := service.Delete(ctx, token, tenantID, hypervisorID, name); err != nil {
		graphql.WriteError(w, err, "delete libvirt network")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteLibvirtNetwork": true,
	})
}
