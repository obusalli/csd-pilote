package pods

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
	graphql.RegisterQuery("pods", "List pods", "csd-pilote.pods.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListPods(ctx, w, variables, service)
		})

	graphql.RegisterQuery("pod", "Get a pod", "csd-pilote.pods.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetPod(ctx, w, variables, service)
		})

	graphql.RegisterQuery("podLogs", "Get pod logs", "csd-pilote.pods.logs",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetPodLogs(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("deletePod", "Delete a pod", "csd-pilote.pods.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeletePod(ctx, w, variables, service)
		})
}

func handleListPods(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterIDStr, ok := variables["clusterId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("clusterId is required"))
		return
	}

	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid clusterId"))
		return
	}

	namespace, _ := variables["namespace"].(string)

	var filter *PodFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &PodFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if phase, ok := f["phase"].(string); ok {
			filter.Phase = &phase
		}
	}

	pods, err := service.List(ctx, token, tenantID, clusterID, namespace, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"pods":      pods,
		"podsCount": len(pods),
	}))
}

func handleGetPod(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterIDStr, ok := variables["clusterId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("clusterId is required"))
		return
	}

	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid clusterId"))
		return
	}

	namespace, ok := variables["namespace"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("namespace is required"))
		return
	}

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	pod, err := service.Get(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"pod": pod,
	}))
}

func handleGetPodLogs(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterIDStr, ok := variables["clusterId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("clusterId is required"))
		return
	}

	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid clusterId"))
		return
	}

	namespace, ok := variables["namespace"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("namespace is required"))
		return
	}

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	container, _ := variables["container"].(string)

	var tailLines int64 = 100
	if tl, ok := variables["tailLines"].(float64); ok {
		tailLines = int64(tl)
	}

	previous := false
	if p, ok := variables["previous"].(bool); ok {
		previous = p
	}

	logs, err := service.GetLogs(ctx, token, tenantID, clusterID, namespace, name, container, tailLines, previous)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"podLogs": logs,
	}))
}

func handleDeletePod(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterIDStr, ok := variables["clusterId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("clusterId is required"))
		return
	}

	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid clusterId"))
		return
	}

	namespace, ok := variables["namespace"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("namespace is required"))
		return
	}

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	var gracePeriod int64 = -1
	if gp, ok := variables["gracePeriod"].(float64); ok {
		gracePeriod = int64(gp)
	}

	if err := service.Delete(ctx, token, tenantID, clusterID, namespace, name, gracePeriod); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deletePod": true,
	}))
}
