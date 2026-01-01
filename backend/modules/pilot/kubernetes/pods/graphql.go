package pods

import (
	"context"
	"net/http"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
)

// Pod phase values for validation
var podPhaseValues = []string{"Pending", "Running", "Succeeded", "Failed", "Unknown", ""}

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
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterID, err := graphql.ParseUUID(variables, "clusterId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Namespace is optional for listing pods
	namespace, _ := variables["namespace"].(string)
	if namespace != "" {
		v := validation.NewValidator()
		v.KubernetesName("namespace", namespace)
		if v.HasErrors() {
			graphql.WriteValidationError(w, v.FirstError())
			return
		}
	}

	var filter *PodFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &PodFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if phase, ok := f["phase"].(string); ok {
			if err := graphql.ValidateEnum(phase, podPhaseValues, "phase"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			filter.Phase = &phase
		}
	}

	pods, err := service.List(ctx, token, tenantID, clusterID, namespace, filter)
	if err != nil {
		graphql.WriteError(w, err, "list pods")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"pods":      pods,
		"podsCount": len(pods),
	})
}

func handleGetPod(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterID, err := graphql.ParseUUID(variables, "clusterId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	namespace, err := graphql.ParseStringRequired(variables, "namespace")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate Kubernetes names (RFC 1123)
	v := validation.NewValidator()
	v.KubernetesName("namespace", namespace)
	v.KubernetesName("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	pod, err := service.Get(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		graphql.WriteError(w, err, "get pod")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"pod": pod,
	})
}

func handleGetPodLogs(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterID, err := graphql.ParseUUID(variables, "clusterId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	namespace, err := graphql.ParseStringRequired(variables, "namespace")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate Kubernetes names (RFC 1123)
	v := validation.NewValidator()
	v.KubernetesName("namespace", namespace)
	v.KubernetesName("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	// Container name is optional
	container, _ := variables["container"].(string)
	if container != "" {
		v.KubernetesName("container", container)
		if v.HasErrors() {
			graphql.WriteValidationError(w, v.FirstError())
			return
		}
	}

	// Limit tailLines to prevent abuse (max 10000 lines)
	var tailLines int64 = 100
	if tl, ok := variables["tailLines"].(float64); ok {
		tailLines = int64(tl)
		if tailLines < 1 {
			tailLines = 1
		}
		if tailLines > 10000 {
			graphql.WriteValidationError(w, "tailLines cannot exceed 10000")
			return
		}
	}

	previous := false
	if p, ok := variables["previous"].(bool); ok {
		previous = p
	}

	logs, err := service.GetLogs(ctx, token, tenantID, clusterID, namespace, name, container, tailLines, previous)
	if err != nil {
		graphql.WriteError(w, err, "get pod logs")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"podLogs": logs,
	})
}

func handleDeletePod(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterID, err := graphql.ParseUUID(variables, "clusterId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	namespace, err := graphql.ParseStringRequired(variables, "namespace")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate Kubernetes names (RFC 1123)
	v := validation.NewValidator()
	v.KubernetesName("namespace", namespace)
	v.KubernetesName("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	// Limit gracePeriod to reasonable value (max 5 minutes)
	var gracePeriod int64 = -1
	if gp, ok := variables["gracePeriod"].(float64); ok {
		gracePeriod = int64(gp)
		if gracePeriod > 300 {
			graphql.WriteValidationError(w, "gracePeriod cannot exceed 300 seconds")
			return
		}
	}

	if err := service.Delete(ctx, token, tenantID, clusterID, namespace, name, gracePeriod); err != nil {
		graphql.WriteError(w, err, "delete pod")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DELETE_POD",
		ResourceType: "k8s_pod",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deletePod": true,
	})
}
