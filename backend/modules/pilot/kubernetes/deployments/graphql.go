package deployments

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
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	clusterID, err := graphql.ParseUUID(variables, "clusterId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Namespace is optional for listing deployments
	namespace, _ := variables["namespace"].(string)
	if namespace != "" {
		v := validation.NewValidator()
		v.KubernetesName("namespace", namespace)
		if v.HasErrors() {
			graphql.WriteValidationError(w, v.FirstError())
			return
		}
	}

	var filter *DeploymentFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &DeploymentFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
	}

	deployments, err := service.List(ctx, token, tenantID, clusterID, namespace, filter)
	if err != nil {
		graphql.WriteError(w, err, "list deployments")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deployments":      deployments,
		"deploymentsCount": len(deployments),
	})
}

func handleGetDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	deployment, err := service.Get(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		graphql.WriteError(w, err, "get deployment")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"deployment": deployment,
	})
}

func handleCreateDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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
	input := &CreateDeploymentInput{}

	if name, ok := inputRaw["name"].(string); ok {
		v.KubernetesName("name", name)
		input.Name = name
	}
	if namespace, ok := inputRaw["namespace"].(string); ok {
		v.KubernetesName("namespace", namespace)
		input.Namespace = namespace
	}
	if image, ok := inputRaw["image"].(string); ok {
		// Validate image reference (max length 1024, safe string)
		v.MaxLength("image", image, 1024).SafeString("image", image)
		input.Image = image
	}
	if replicas, ok := inputRaw["replicas"].(float64); ok {
		replicasInt := int32(replicas)
		// Limit replicas to reasonable value (0-1000)
		if replicasInt < 0 || replicasInt > 1000 {
			graphql.WriteValidationError(w, "replicas must be between 0 and 1000")
			return
		}
		input.Replicas = replicasInt
	}

	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	if input.Name == "" || input.Namespace == "" || input.Image == "" {
		graphql.WriteValidationError(w, "name, namespace and image are required")
		return
	}

	deployment, err := service.Create(ctx, token, tenantID, clusterID, input)
	if err != nil {
		graphql.WriteError(w, err, "create deployment")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "CREATE_DEPLOYMENT",
		ResourceType: "k8s_deployment",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name":      deployment.Name,
			"namespace": deployment.Namespace,
			"image":     input.Image,
			"replicas":  input.Replicas,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"createDeployment": deployment,
	})
}

func handleScaleDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	replicasFloat, ok := variables["replicas"].(float64)
	if !ok {
		graphql.WriteValidationError(w, "replicas is required")
		return
	}

	replicas := int32(replicasFloat)
	// Limit replicas to reasonable value (0-1000)
	if replicas < 0 || replicas > 1000 {
		graphql.WriteValidationError(w, "replicas must be between 0 and 1000")
		return
	}

	deployment, err := service.Scale(ctx, token, tenantID, clusterID, namespace, name, replicas)
	if err != nil {
		graphql.WriteError(w, err, "scale deployment")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "SCALE_DEPLOYMENT",
		ResourceType: "k8s_deployment",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name":      name,
			"namespace": namespace,
			"replicas":  replicas,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"scaleDeployment": deployment,
	})
}

func handleRestartDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	deployment, err := service.Restart(ctx, token, tenantID, clusterID, namespace, name)
	if err != nil {
		graphql.WriteError(w, err, "restart deployment")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "RESTART_DEPLOYMENT",
		ResourceType: "k8s_deployment",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"restartDeployment": deployment,
	})
}

func handleDeleteDeployment(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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
		graphql.WriteError(w, err, "delete deployment")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DELETE_DEPLOYMENT",
		ResourceType: "k8s_deployment",
		ResourceID:   clusterID.String(),
		Details: map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteDeployment": true,
	})
}
