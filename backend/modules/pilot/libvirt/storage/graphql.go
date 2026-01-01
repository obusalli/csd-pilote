package storage

import (
	"context"
	"net/http"

	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
)

// Volume format values for validation
var volumeFormatValues = []string{"raw", "qcow2", "qcow", "vmdk", "vdi", ""}

func init() {
	service := NewService()

	// Storage Pool Queries
	graphql.RegisterQuery("storagePools", "List storage pools", "csd-pilote.storage.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListPools(ctx, w, variables, service)
		})

	graphql.RegisterQuery("storagePool", "Get a storage pool", "csd-pilote.storage.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetPool(ctx, w, variables, service)
		})

	// Storage Pool Mutations
	graphql.RegisterMutation("startStoragePool", "Start a storage pool", "csd-pilote.storage.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleStartPool(ctx, w, variables, service)
		})

	graphql.RegisterMutation("stopStoragePool", "Stop a storage pool", "csd-pilote.storage.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleStopPool(ctx, w, variables, service)
		})

	graphql.RegisterMutation("refreshStoragePool", "Refresh a storage pool", "csd-pilote.storage.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleRefreshPool(ctx, w, variables, service)
		})

	// Volume Queries
	graphql.RegisterQuery("storageVolumes", "List storage volumes", "csd-pilote.storage.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListVolumes(ctx, w, variables, service)
		})

	graphql.RegisterQuery("storageVolume", "Get a storage volume", "csd-pilote.storage.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetVolume(ctx, w, variables, service)
		})

	// Volume Mutations
	graphql.RegisterMutation("createStorageVolume", "Create a storage volume", "csd-pilote.storage.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateVolume(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteStorageVolume", "Delete a storage volume", "csd-pilote.storage.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteVolume(ctx, w, variables, service)
		})
}

func handleListPools(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	var filter *StoragePoolFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &StoragePoolFilter{}
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

	pools, err := service.ListPools(ctx, token, tenantID, hypervisorID, filter)
	if err != nil {
		graphql.WriteError(w, err, "list storage pools")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"storagePools":      pools,
		"storagePoolsCount": len(pools),
	})
}

func handleGetPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.GetPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		graphql.WriteError(w, err, "get storage pool")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"storagePool": pool,
	})
}

func handleStartPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.StartPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		graphql.WriteError(w, err, "start storage pool")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"startStoragePool": pool,
	})
}

func handleStopPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.StopPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		graphql.WriteError(w, err, "stop storage pool")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"stopStoragePool": pool,
	})
}

func handleRefreshPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.RefreshPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		graphql.WriteError(w, err, "refresh storage pool")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"refreshStoragePool": pool,
	})
}

func handleListVolumes(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, err := graphql.ParseStringRequired(variables, "poolName")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.MaxLength("poolName", poolName, validation.MaxNameLength).SafeString("poolName", poolName)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	var filter *StorageVolumeFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &StorageVolumeFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
	}

	volumes, err := service.ListVolumes(ctx, token, tenantID, hypervisorID, poolName, filter)
	if err != nil {
		graphql.WriteError(w, err, "list storage volumes")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"storageVolumes":      volumes,
		"storageVolumesCount": len(volumes),
	})
}

func handleGetVolume(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, err := graphql.ParseStringRequired(variables, "poolName")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	volumeName, err := graphql.ParseStringRequired(variables, "volumeName")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.MaxLength("poolName", poolName, validation.MaxNameLength).SafeString("poolName", poolName)
	v.MaxLength("volumeName", volumeName, validation.MaxNameLength).SafeString("volumeName", volumeName)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	volume, err := service.GetVolume(ctx, token, tenantID, hypervisorID, poolName, volumeName)
	if err != nil {
		graphql.WriteError(w, err, "get storage volume")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"storageVolume": volume,
	})
}

func handleCreateVolume(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, err := graphql.ParseStringRequired(variables, "poolName")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	v := validation.NewValidator()
	v.MaxLength("poolName", poolName, validation.MaxNameLength).SafeString("poolName", poolName)

	input := &CreateVolumeInput{}
	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		v.Required("name", name)
		input.Name = name
	}
	if capacity, ok := inputRaw["capacity"].(float64); ok {
		cap := uint64(capacity)
		// Limit capacity to 10TB
		if cap > 10*1024*1024*1024*1024 {
			graphql.WriteValidationError(w, "capacity exceeds maximum allowed (10TB)")
			return
		}
		input.Capacity = cap
	}
	if format, ok := inputRaw["format"].(string); ok {
		if err := graphql.ValidateEnum(format, volumeFormatValues, "format"); err != nil {
			graphql.WriteValidationError(w, err.Error())
			return
		}
		input.Format = format
	}

	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	if input.Name == "" || input.Capacity == 0 {
		graphql.WriteValidationError(w, "name and capacity are required")
		return
	}

	volume, err := service.CreateVolume(ctx, token, tenantID, hypervisorID, poolName, input)
	if err != nil {
		graphql.WriteError(w, err, "create storage volume")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"createStorageVolume": volume,
	})
}

func handleDeleteVolume(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, err := graphql.ParseStringRequired(variables, "poolName")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	volumeName, err := graphql.ParseStringRequired(variables, "volumeName")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	v := validation.NewValidator()
	v.MaxLength("poolName", poolName, validation.MaxNameLength).SafeString("poolName", poolName)
	v.MaxLength("volumeName", volumeName, validation.MaxNameLength).SafeString("volumeName", volumeName)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	if err := service.DeleteVolume(ctx, token, tenantID, hypervisorID, poolName, volumeName); err != nil {
		graphql.WriteError(w, err, "delete storage volume")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteStorageVolume": true,
	})
}
