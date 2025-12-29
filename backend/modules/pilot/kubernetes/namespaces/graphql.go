package namespaces

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
	graphql.RegisterQuery("namespaces", "List namespaces for a cluster", "csd-pilote.namespaces.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListNamespaces(ctx, w, variables, service)
		})

	graphql.RegisterQuery("namespace", "Get a namespace by name", "csd-pilote.namespaces.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetNamespace(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("createNamespace", "Create a namespace", "csd-pilote.namespaces.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateNamespace(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteNamespace", "Delete a namespace", "csd-pilote.namespaces.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteNamespace(ctx, w, variables, service)
		})
}

func handleListNamespaces(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	var filter *NamespaceFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &NamespaceFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
	}

	namespaces, err := service.List(ctx, token, tenantID, clusterID, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"namespaces":      namespaces,
		"namespacesCount": len(namespaces),
	}))
}

func handleGetNamespace(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	namespace, err := service.Get(ctx, token, tenantID, clusterID, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"namespace": namespace,
	}))
}

func handleCreateNamespace(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	var labels map[string]string
	if l, ok := variables["labels"].(map[string]interface{}); ok {
		labels = make(map[string]string)
		for k, v := range l {
			if s, ok := v.(string); ok {
				labels[k] = s
			}
		}
	}

	namespace, err := service.Create(ctx, token, tenantID, clusterID, name, labels)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createNamespace": namespace,
	}))
}

func handleDeleteNamespace(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, ok := variables["name"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}

	if err := service.Delete(ctx, token, tenantID, clusterID, name); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteNamespace": true,
	}))
}
