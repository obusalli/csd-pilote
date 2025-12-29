package storage

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

	var filter *StoragePoolFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &StoragePoolFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if active, ok := f["active"].(bool); ok {
			filter.Active = &active
		}
	}

	pools, err := service.ListPools(ctx, token, tenantID, hypervisorID, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"storagePools":      pools,
		"storagePoolsCount": len(pools),
	}))
}

func handleGetPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.GetPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"storagePool": pool,
	}))
}

func handleStartPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.StartPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"startStoragePool": pool,
	}))
}

func handleStopPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.StopPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"stopStoragePool": pool,
	}))
}

func handleRefreshPool(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	pool, err := service.RefreshPool(ctx, token, tenantID, hypervisorID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"refreshStoragePool": pool,
	}))
}

func handleListVolumes(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, ok := variables["poolName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("poolName is required"))
		return
	}

	var filter *StorageVolumeFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &StorageVolumeFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
	}

	volumes, err := service.ListVolumes(ctx, token, tenantID, hypervisorID, poolName, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"storageVolumes":      volumes,
		"storageVolumesCount": len(volumes),
	}))
}

func handleGetVolume(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, ok := variables["poolName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("poolName is required"))
		return
	}

	volumeName, ok := variables["volumeName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("volumeName is required"))
		return
	}

	volume, err := service.GetVolume(ctx, token, tenantID, hypervisorID, poolName, volumeName)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"storageVolume": volume,
	}))
}

func handleCreateVolume(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, ok := variables["poolName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("poolName is required"))
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := &CreateVolumeInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if capacity, ok := inputRaw["capacity"].(float64); ok {
		input.Capacity = uint64(capacity)
	}
	if format, ok := inputRaw["format"].(string); ok {
		input.Format = format
	}

	if input.Name == "" || input.Capacity == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name and capacity are required"))
		return
	}

	volume, err := service.CreateVolume(ctx, token, tenantID, hypervisorID, poolName, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createStorageVolume": volume,
	}))
}

func handleDeleteVolume(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	poolName, ok := variables["poolName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("poolName is required"))
		return
	}

	volumeName, ok := variables["volumeName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("volumeName is required"))
		return
	}

	if err := service.DeleteVolume(ctx, token, tenantID, hypervisorID, poolName, volumeName); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteStorageVolume": true,
	}))
}
