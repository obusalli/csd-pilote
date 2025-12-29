package deployments

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
	graphql.RegisterQuery("deployments", "List deployments", "csd-pilote.deployments.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListDeployments(ctx, w, variables, service)
		})

	graphql.RegisterQuery("deployment", "Get a deployment", "csd-pilote.deployments.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetDeployment(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("createDeployment", "Create a deployment", "csd-pilote.deployments.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateDeployment(ctx, w, variables, service)
		})

	graphql.RegisterMutation("scaleDeployment", "Scale a deployment", "csd-pilote.deployments.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleScaleDeployment(ctx, w, variables, service)
		})

	graphql.RegisterMutation("restartDeployment", "Restart a deployment", "csd-pilote.deployments.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleRestartDeployment(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteDeployment", "Delete a deployment", "csd-pilote.deployments.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteDeployment(ctx, w, variables, service)
		})
}

func handleListDeployments(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	var filter *DeploymentFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &DeploymentFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
	}

	deployments, err := service.List(ctx, token, tenantID, clusterID, namespace, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deployments":      deployments,
		"deploymentsCount": len(deployments),
	}))
}

func handleGetDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	deployment, err := service.Get(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deployment": deployment,
	}))
}

func handleCreateDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := &CreateDeploymentInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if namespace, ok := inputRaw["namespace"].(string); ok {
		input.Namespace = namespace
	}
	if image, ok := inputRaw["image"].(string); ok {
		input.Image = image
	}
	if replicas, ok := inputRaw["replicas"].(float64); ok {
		input.Replicas = int32(replicas)
	}

	if input.Name == "" || input.Namespace == "" || input.Image == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name, namespace and image are required"))
		return
	}

	deployment, err := service.Create(ctx, token, tenantID, clusterID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createDeployment": deployment,
	}))
}

func handleScaleDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	replicasFloat, ok := variables["replicas"].(float64)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("replicas is required"))
		return
	}

	deployment, err := service.Scale(ctx, token, tenantID, clusterID, namespace, name, int32(replicasFloat))
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"scaleDeployment": deployment,
	}))
}

func handleRestartDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	deployment, err := service.Restart(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"restartDeployment": deployment,
	}))
}

func handleDeleteDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	if err := service.Delete(ctx, token, tenantID, clusterID, namespace, name); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteDeployment": true,
	}))
}
