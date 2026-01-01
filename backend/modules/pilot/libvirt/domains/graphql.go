package domains

import (
	"context"
	"net/http"

	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
)

// Domain state enum values for validation
var domainStateValues = []string{"running", "paused", "shutoff", "crashed", "suspended", ""}

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
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	hypervisorID, err := graphql.ParseUUID(variables, "hypervisorId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	var filter *DomainFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &DomainFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if state, ok := f["state"].(string); ok {
			if err := graphql.ValidateEnum(state, domainStateValues, "state"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			s := DomainState(state)
			filter.State = &s
		}
	}

	domains, err := service.List(ctx, token, tenantID, hypervisorID, filter)
	if err != nil {
		graphql.WriteError(w, err, "list domains")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"domains":      domains,
		"domainsCount": len(domains),
	})
}

func handleGetDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, err := graphql.ParseStringRequired(variables, "uuid")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate UUID format
	v := validation.NewValidator()
	v.UUID("uuid", domainUUID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	domain, err := service.Get(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		graphql.WriteError(w, err, "get domain")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"domain": domain,
	})
}

func handleStartDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, err := graphql.ParseStringRequired(variables, "uuid")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.UUID("uuid", domainUUID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	domain, err := service.Start(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		graphql.WriteError(w, err, "start domain")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"startDomain": domain,
	})
}

func handleShutdownDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, err := graphql.ParseStringRequired(variables, "uuid")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.UUID("uuid", domainUUID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	domain, err := service.Shutdown(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		graphql.WriteError(w, err, "shutdown domain")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"shutdownDomain": domain,
	})
}

func handleForceStopDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, err := graphql.ParseStringRequired(variables, "uuid")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.UUID("uuid", domainUUID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	domain, err := service.ForceStop(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		graphql.WriteError(w, err, "force stop domain")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"forceStopDomain": domain,
	})
}

func handleRebootDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, err := graphql.ParseStringRequired(variables, "uuid")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.UUID("uuid", domainUUID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	domain, err := service.Reboot(ctx, token, tenantID, hypervisorID, domainUUID)
	if err != nil {
		graphql.WriteError(w, err, "reboot domain")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"rebootDomain": domain,
	})
}

func handleDeleteDomain(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, err := graphql.ParseStringRequired(variables, "uuid")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.UUID("uuid", domainUUID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	deleteVolumes := graphql.ParseBool(variables, "deleteVolumes", false)

	if err := service.Delete(ctx, token, tenantID, hypervisorID, domainUUID, deleteVolumes); err != nil {
		graphql.WriteError(w, err, "delete domain")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteDomain": true,
	})
}

func handleSetDomainAutostart(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	domainUUID, err := graphql.ParseStringRequired(variables, "uuid")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.UUID("uuid", domainUUID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	autostart, ok := variables["autostart"].(bool)
	if !ok {
		graphql.WriteValidationError(w, "autostart is required")
		return
	}

	domain, err := service.SetAutostart(ctx, token, tenantID, hypervisorID, domainUUID, autostart)
	if err != nil {
		graphql.WriteError(w, err, "set domain autostart")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"setDomainAutostart": domain,
	})
}
