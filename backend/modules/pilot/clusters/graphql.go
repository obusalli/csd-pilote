package clusters

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
	graphql.RegisterQuery("clusters", "List all clusters", "csd-pilote.clusters.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListClusters(ctx, w, variables, service)
		})

	graphql.RegisterQuery("cluster", "Get a cluster by ID", "csd-pilote.clusters.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetCluster(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("createCluster", "Create a new cluster", "csd-pilote.clusters.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateCluster(ctx, w, variables, service)
		})

	graphql.RegisterMutation("updateCluster", "Update a cluster", "csd-pilote.clusters.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleUpdateCluster(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteCluster", "Delete a cluster", "csd-pilote.clusters.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteCluster(ctx, w, variables, service)
		})

	graphql.RegisterMutation("testClusterConnection", "Test cluster connection", "csd-pilote.clusters.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleTestClusterConnection(ctx, w, variables, service)
		})
}

func handleListClusters(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	var filter *ClusterFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &ClusterFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if status, ok := f["status"].(string); ok {
			s := ClusterStatus(status)
			filter.Status = &s
		}
	}

	clusters, count, err := service.List(ctx, tenantID, filter, limit, offset)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"clusters":      clusters,
		"clustersCount": count,
	}))
}

func handleGetCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	cluster, err := service.Get(ctx, tenantID, id)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"cluster": cluster,
	}))
}

func handleCreateCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input := &ClusterInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}
	if input.ArtifactKey == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("artifactKey is required"))
		return
	}

	cluster, err := service.Create(ctx, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createCluster": cluster,
	}))
}

func handleUpdateCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input := &ClusterInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}

	cluster, err := service.Update(ctx, tenantID, id, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"updateCluster": cluster,
	}))
}

func handleDeleteCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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
		"deleteCluster": true,
	}))
}

func handleTestClusterConnection(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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
		"testClusterConnection": true,
	}))
}
