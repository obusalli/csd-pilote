package domains

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
	graphql.RegisterQuery("domains", "List domains (VMs)", "csd-pilote.domains.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListDomains(ctx, w, variables, service)
		})

	graphql.RegisterQuery("domain", "Get a domain", "csd-pilote.domains.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetDomain(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("startDomain", "Start a domain", "csd-pilote.domains.power",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleStartDomain(ctx, w, variables, service)
		})

	graphql.RegisterMutation("shutdownDomain", "Shutdown a domain", "csd-pilote.domains.power",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleShutdownDomain(ctx, w, variables, service)
		})

	graphql.RegisterMutation("forceStopDomain", "Force stop a domain", "csd-pilote.domains.power",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleForceStopDomain(ctx, w, variables, service)
		})

	graphql.RegisterMutation("rebootDomain", "Reboot a domain", "csd-pilote.domains.power",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleRebootDomain(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteDomain", "Delete a domain", "csd-pilote.domains.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteDomain(ctx, w, variables, service)
		})

	graphql.RegisterMutation("setDomainAutostart", "Set domain autostart", "csd-pilote.domains.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleSetDomainAutostart(ctx, w, variables, service)
		})
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

	var filter *DomainFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &DomainFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if state, ok := f["state"].(string); ok {
			s := DomainState(state)
			filter.State = &s
		}
	}

	domains, err := service.List(ctx, token, tenantID, hypervisorID, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"domains":      domains,
		"domainsCount": len(domains),
	}))
}

func handleGetDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, ok := variables["uuid"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uuid is required"))
		return
	}

	domain, err := service.Get(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"domain": domain,
	}))
}

func handleStartDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, ok := variables["uuid"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uuid is required"))
		return
	}

	domain, err := service.Start(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"startDomain": domain,
	}))
}

func handleShutdownDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, ok := variables["uuid"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uuid is required"))
		return
	}

	domain, err := service.Shutdown(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"shutdownDomain": domain,
	}))
}

func handleForceStopDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, ok := variables["uuid"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uuid is required"))
		return
	}

	domain, err := service.ForceStop(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"forceStopDomain": domain,
	}))
}

func handleRebootDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, ok := variables["uuid"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uuid is required"))
		return
	}

	domain, err := service.Reboot(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"rebootDomain": domain,
	}))
}

func handleDeleteDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, ok := variables["uuid"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uuid is required"))
		return
	}

	deleteVolumes := false
	if dv, ok := variables["deleteVolumes"].(bool); ok {
		deleteVolumes = dv
	}

	if err := service.Delete(ctx, token, tenantID, hypervisorID, domainUUID, deleteVolumes); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteDomain": true,
	}))
}

func handleSetDomainAutostart(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, ok := variables["uuid"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("uuid is required"))
		return
	}

	autostart, ok := variables["autostart"].(bool)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("autostart is required"))
		return
	}

	domain, err := service.SetAutostart(ctx, token, tenantID, hypervisorID, domainUUID, autostart)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"setDomainAutostart": domain,
	}))
}
