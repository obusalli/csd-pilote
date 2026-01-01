package namespaces

import (
	"context"
	"net/http"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
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
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterID, err := graphql.ParseUUID(variables, "clusterId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	var filter *NamespaceFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &NamespaceFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
	}

	namespaces, err := service.List(ctx, token, tenantID, clusterID, filter)
	if err != nil {
		graphql.WriteError(w, err, "list namespaces")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"namespaces":      namespaces,
		"namespacesCount": len(namespaces),
	})
}

func handleGetNamespace(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate Kubernetes name (RFC 1123)
	v := validation.NewValidator()
	v.KubernetesName("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	namespace, err := service.Get(ctx, token, tenantID, clusterID, name)
	if err != nil {
		graphql.WriteError(w, err, "get namespace")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"namespace": namespace,
	})
}

func handleCreateNamespace(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate Kubernetes name
	v := validation.NewValidator()
	v.KubernetesName("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	var labels map[string]string
	if l, ok := variables["labels"].(map[string]interface{}); ok {
		// Limit number of labels
		if len(l) > 50 {
			graphql.WriteValidationError(w, "too many labels (max 50)")
			return
		}
		labels = make(map[string]string)
		for k, val := range l {
			if s, ok := val.(string); ok {
				// Validate label key and value
				v.MaxLength("label key", k, 253).SafeString("label key", k)
				v.MaxLength("label value", s, 63).SafeString("label value", s)
				labels[k] = s
			}
		}
		if v.HasErrors() {
			graphql.WriteValidationError(w, v.FirstError())
			return
		}
	}

	namespace, err := service.Create(ctx, token, tenantID, clusterID, name, labels)
	if err != nil {
		graphql.WriteError(w, err, "create namespace")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "CREATE_NAMESPACE",
		ResourceType: "k8s_namespace",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name": namespace.Name,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"createNamespace": namespace,
	})
}

func handleDeleteNamespace(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	name, err := graphql.ParseStringRequired(variables, "name")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate Kubernetes name
	v := validation.NewValidator()
	v.KubernetesName("name", name)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	if err := service.Delete(ctx, token, tenantID, clusterID, name); err != nil {
		graphql.WriteError(w, err, "delete namespace")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DELETE_NAMESPACE",
		ResourceType: "k8s_namespace",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name": name,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteNamespace": true,
	})
}
