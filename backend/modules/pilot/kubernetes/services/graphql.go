package services

import (
	"context"
	"net/http"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
)

// Kubernetes service type values for validation
var serviceTypeValues = []string{"ClusterIP", "NodePort", "LoadBalancer", "ExternalName", ""}

// Protocol values for validation
var protocolValues = []string{"TCP", "UDP", "SCTP", ""}

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
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterID, err := graphql.ParseUUID(variables, "clusterId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Namespace is optional for listing services
	namespace, _ := variables["namespace"].(string)
	if namespace != "" {
		v := validation.NewValidator()
		v.KubernetesName("namespace", namespace)
		if v.HasErrors() {
			graphql.WriteValidationError(w, v.FirstError())
			return
		}
	}

	var filter *ServiceFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &ServiceFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if svcType, ok := f["type"].(string); ok {
			if err := graphql.ValidateEnum(svcType, serviceTypeValues, "type"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			filter.Type = &svcType
		}
	}

	services, err := service.List(ctx, token, tenantID, clusterID, namespace, filter)
	if err != nil {
		graphql.WriteError(w, err, "list k8s services")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"k8sServices":      services,
		"k8sServicesCount": len(services),
	})
}

func handleGetService(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	svc, err := service.Get(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		graphql.WriteError(w, err, "get k8s service")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"k8sService": svc,
	})
}

func handleCreateService(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	v := validation.NewValidator()
	input := &CreateServiceInput{}

	if name, ok := inputRaw["name"].(string); ok {
		v.KubernetesName("name", name)
		input.Name = name
	}
	if namespace, ok := inputRaw["namespace"].(string); ok {
		v.KubernetesName("namespace", namespace)
		input.Namespace = namespace
	}
	if svcType, ok := inputRaw["type"].(string); ok {
		if err := graphql.ValidateEnum(svcType, serviceTypeValues, "type"); err != nil {
			graphql.WriteValidationError(w, err.Error())
			return
		}
		input.Type = svcType
	}

	// Validate selector (limit to 50 entries)
	if selector, ok := inputRaw["selector"].(map[string]interface{}); ok {
		if len(selector) > 50 {
			graphql.WriteValidationError(w, "too many selector entries (max 50)")
			return
		}
		input.Selector = make(map[string]string)
		for k, val := range selector {
			if s, ok := val.(string); ok {
				v.MaxLength("selector key", k, 253).SafeString("selector key", k)
				v.MaxLength("selector value", s, 63).SafeString("selector value", s)
				input.Selector[k] = s
			}
		}
	}

	// Validate ports (limit to 20 ports)
	if portsRaw, ok := inputRaw["ports"].([]interface{}); ok {
		if len(portsRaw) > 20 {
			graphql.WriteValidationError(w, "too many ports (max 20)")
			return
		}
		input.Ports = make([]ServicePortInput, len(portsRaw))
		for i, pRaw := range portsRaw {
			if p, ok := pRaw.(map[string]interface{}); ok {
				if name, ok := p["name"].(string); ok {
					v.MaxLength("port name", name, 63).SafeString("port name", name)
					input.Ports[i].Name = name
				}
				if protocol, ok := p["protocol"].(string); ok {
					if err := graphql.ValidateEnum(protocol, protocolValues, "protocol"); err != nil {
						graphql.WriteValidationError(w, err.Error())
						return
					}
					input.Ports[i].Protocol = protocol
				}
				if port, ok := p["port"].(float64); ok {
					portInt := int32(port)
					if portInt < 1 || portInt > 65535 {
						graphql.WriteValidationError(w, "port must be between 1 and 65535")
						return
					}
					input.Ports[i].Port = portInt
				}
				if targetPort, ok := p["targetPort"].(float64); ok {
					targetPortInt := int32(targetPort)
					if targetPortInt < 1 || targetPortInt > 65535 {
						graphql.WriteValidationError(w, "targetPort must be between 1 and 65535")
						return
					}
					input.Ports[i].TargetPort = targetPortInt
				}
			}
		}
	}

	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	if input.Name == "" || input.Namespace == "" || len(input.Ports) == 0 {
		graphql.WriteValidationError(w, "name, namespace and ports are required")
		return
	}

	svc, err := service.Create(ctx, token, tenantID, clusterID, input)
	if err != nil {
		graphql.WriteError(w, err, "create k8s service")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "CREATE_K8S_SERVICE",
		ResourceType: "k8s_service",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name":      svc.Name,
			"namespace": svc.Namespace,
			"type":      input.Type,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"createK8sService": svc,
	})
}

func handleDeleteService(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	if err := service.Delete(ctx, token, tenantID, clusterID, namespace, name); err != nil {
		graphql.WriteError(w, err, "delete k8s service")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DELETE_K8S_SERVICE",
		ResourceType: "k8s_service",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteK8sService": true,
	})
}
