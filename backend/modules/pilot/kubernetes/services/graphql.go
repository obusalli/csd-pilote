package services

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
	graphql.RegisterQuery("k8sServices", "List Kubernetes services", "csd-pilote.services.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListServices(ctx, w, variables, service)
		})

	graphql.RegisterQuery("k8sService", "Get a Kubernetes service", "csd-pilote.services.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetService(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("createK8sService", "Create a Kubernetes service", "csd-pilote.services.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateService(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteK8sService", "Delete a Kubernetes service", "csd-pilote.services.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteService(ctx, w, variables, service)
		})
}

func handleListServices(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	var filter *ServiceFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &ServiceFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if svcType, ok := f["type"].(string); ok {
			filter.Type = &svcType
		}
	}

	services, err := service.List(ctx, token, tenantID, clusterID, namespace, filter)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"k8sServices":      services,
		"k8sServicesCount": len(services),
	}))
}

func handleGetService(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	svc, err := service.Get(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"k8sService": svc,
	}))
}

func handleCreateService(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input := &CreateServiceInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if namespace, ok := inputRaw["namespace"].(string); ok {
		input.Namespace = namespace
	}
	if svcType, ok := inputRaw["type"].(string); ok {
		input.Type = svcType
	}
	if selector, ok := inputRaw["selector"].(map[string]interface{}); ok {
		input.Selector = make(map[string]string)
		for k, v := range selector {
			if s, ok := v.(string); ok {
				input.Selector[k] = s
			}
		}
	}
	if portsRaw, ok := inputRaw["ports"].([]interface{}); ok {
		input.Ports = make([]ServicePortInput, len(portsRaw))
		for i, pRaw := range portsRaw {
			if p, ok := pRaw.(map[string]interface{}); ok {
				if name, ok := p["name"].(string); ok {
					input.Ports[i].Name = name
				}
				if protocol, ok := p["protocol"].(string); ok {
					input.Ports[i].Protocol = protocol
				}
				if port, ok := p["port"].(float64); ok {
					input.Ports[i].Port = int32(port)
				}
				if targetPort, ok := p["targetPort"].(float64); ok {
					input.Ports[i].TargetPort = int32(targetPort)
				}
			}
		}
	}

	if input.Name == "" || input.Namespace == "" || len(input.Ports) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name, namespace and ports are required"))
		return
	}

	svc, err := service.Create(ctx, token, tenantID, clusterID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createK8sService": svc,
	}))
}

func handleDeleteService(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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
		"deleteK8sService": true,
	}))
}
