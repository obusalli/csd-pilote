package clusters

import (
	"context"
	"net/http"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/graphql/crud"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
)

func init() {
	service := NewService()

	// Queries
	graphql.RegisterQuery("kubernetesAgents", "List agents that support Kubernetes management", "csd-pilote.clusters.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListKubernetesAgents(ctx, w, variables, service)
		})

	graphql.RegisterQuery("kubernetesDeployAgents", "List agents that can deploy Kubernetes clusters", "csd-pilote.clusters.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListKubernetesDeployAgents(ctx, w, variables, service)
		})

	graphql.RegisterQuery("kubernetesDistributions", "List deployable Kubernetes distributions", "csd-pilote.clusters.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListKubernetesDistributions(ctx, w, variables, service)
		})

	graphql.RegisterQuery("allKubernetesDistributions", "List all Kubernetes distributions (deployable + connect-only)", "csd-pilote.clusters.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListAllKubernetesDistributions(ctx, w, variables, service)
		})

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

	graphql.RegisterMutation("deployCluster", "Deploy a new Kubernetes cluster", "csd-pilote.clusters.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeployCluster(ctx, w, variables, service)
		})

	graphql.RegisterMutation("bulkDeleteClusters", "Delete multiple clusters", "csd-pilote.clusters.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleBulkDeleteClusters(ctx, w, variables, service)
		})
}

func handleListClusters(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	// Use validated pagination with max limits
	limit, offset := graphql.ParsePagination(variables)

	var filter *ClusterFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &ClusterFilter{}
		if search, ok := f["search"].(string); ok {
			// Validate search length
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search query too long")
				return
			}
			filter.Search = &search
		}
		if status, ok := f["status"].(string); ok {
			// Validate enum
			if err := graphql.ValidateEnum(status, graphql.ClusterStatusValues, "status"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			s := ClusterStatus(status)
			filter.Status = &s
		}
		if mode, ok := f["mode"].(string); ok {
			if err := graphql.ValidateEnum(mode, graphql.ClusterModeValues, "mode"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			m := ClusterMode(mode)
			filter.Mode = &m
		}
		if distro, ok := f["distribution"].(string); ok {
			if err := graphql.ValidateEnum(distro, graphql.KubernetesDistroValues, "distribution"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			d := KubernetesDistribution(distro)
			filter.Distribution = &d
		}
	}

	clusters, count, err := service.List(ctx, tenantID, filter, limit, offset)
	if err != nil {
		graphql.WriteError(w, err, "list clusters")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"clusters":      clusters,
		"clustersCount": count,
	})
}

func handleGetCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	hctx := crud.ExtractTenantContext(ctx, w)
	if hctx == nil {
		return
	}

	id, ok := crud.ParseID(w, variables, "id")
	if !ok {
		return
	}

	cluster, err := service.Get(ctx, hctx.TenantID, id)
	if err != nil {
		crud.HandleError(w, err, "get cluster")
		return
	}

	crud.WriteGetResult(w, "cluster", cluster)
}

func handleCreateCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	hctx := crud.ExtractFullContext(ctx, w)
	if hctx == nil {
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		crud.HandleValidationError(w, "input is required")
		return
	}

	input, err := parseClusterInput(inputRaw)
	if err != nil {
		crud.HandleValidationError(w, err.Error())
		return
	}

	// Validate required fields
	v := validation.NewValidator()
	v.Required("name", input.Name).MaxLength("name", input.Name, validation.MaxNameLength)
	v.Required("agentId", input.AgentID).UUID("agentId", input.AgentID)
	v.Required("artifactKey", input.ArtifactKey).MaxLength("artifactKey", input.ArtifactKey, validation.MaxNameLength)
	if input.Description != "" {
		v.MaxLength("description", input.Description, validation.MaxDescriptionLength)
	}
	if v.HasErrors() {
		crud.HandleValidationError(w, v.FirstError())
		return
	}

	cluster, err := service.Create(ctx, hctx.TenantID, hctx.UserID, input)
	if err != nil {
		crud.HandleError(w, err, "create cluster")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, hctx.Token, csdcore.AuditEntry{
		Action:       "CREATE_CLUSTER",
		ResourceType: "cluster",
		ResourceID:   cluster.ID.String(),
		Details: map[string]interface{}{
			"name":         cluster.Name,
			"mode":         cluster.Mode,
			"distribution": cluster.Distribution,
		},
	})

	crud.WriteCreateResult(w, "createCluster", cluster)
}

// parseClusterInput parses and validates cluster input
func parseClusterInput(inputRaw map[string]interface{}) (*ClusterInput, error) {
	input := &ClusterInput{}

	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if agentId, ok := inputRaw["agentId"].(string); ok {
		input.AgentID = agentId
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}
	if distribution, ok := inputRaw["distribution"].(string); ok {
		if err := graphql.ValidateEnum(distribution, graphql.KubernetesDistroValues, "distribution"); err != nil {
			return nil, err
		}
		input.Distribution = KubernetesDistribution(distribution)
	}

	return input, nil
}

func handleUpdateCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseClusterInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate field lengths if provided
	v := validation.NewValidator()
	if input.Name != "" {
		v.MaxLength("name", input.Name, validation.MaxNameLength)
	}
	if input.Description != "" {
		v.MaxLength("description", input.Description, validation.MaxDescriptionLength)
	}
	if input.AgentID != "" {
		v.UUID("agentId", input.AgentID)
	}
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	cluster, err := service.Update(ctx, tenantID, id, input)
	if err != nil {
		graphql.WriteError(w, err, "update cluster")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "UPDATE_CLUSTER",
		ResourceType: "cluster",
		ResourceID:   cluster.ID.String(),
		Details: map[string]interface{}{
			"name": cluster.Name,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"updateCluster": cluster,
	})
}

func handleDeleteCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Get cluster info before deletion for audit
	cluster, _ := service.Get(ctx, tenantID, id)
	clusterName := ""
	if cluster != nil {
		clusterName = cluster.Name
	}

	if err := service.Delete(ctx, tenantID, id); err != nil {
		graphql.WriteError(w, err, "delete cluster")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DELETE_CLUSTER",
		ResourceType: "cluster",
		ResourceID:   id.String(),
		Details: map[string]interface{}{
			"name": clusterName,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteCluster": true,
	})
}

func handleTestClusterConnection(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// agentId is optional - parse it if provided
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	if err := service.TestConnection(ctx, token, tenantID, id, agentID); err != nil {
		graphql.WriteError(w, err, "test cluster connection")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"testClusterConnection": true,
	})
}

func handleListKubernetesAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	client := csdcore.GetClient()
	agents, err := client.ListAgentsByCapability(ctx, token, "kubernetes")
	if err != nil {
		graphql.WriteError(w, err, "list kubernetes agents")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"kubernetesAgents": agents,
	})
}

func handleListKubernetesDeployAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	// Filter by distribution if provided
	distribution, _ := variables["distribution"].(string)

	client := csdcore.GetClient()
	var agents []csdcore.Agent
	var err error

	if distribution != "" {
		// Filter by specific distribution capability (e.g., "kubernetes-deploy-k3s")
		capability := "kubernetes-deploy-" + distribution
		agents, err = client.ListAgentsByCapability(ctx, token, capability)
	} else {
		// Get all agents that can deploy any Kubernetes distribution
		agents, err = client.ListAgentsByCapabilityPrefix(ctx, token, "kubernetes-deploy-")
	}

	if err != nil {
		graphql.WriteError(w, err, "list kubernetes deploy agents")
		return
	}

	// Enrich agents with their supported distributions
	type AgentWithDistros struct {
		csdcore.Agent
		SupportedDistributions []string `json:"supportedDistributions"`
	}

	enrichedAgents := make([]AgentWithDistros, 0, len(agents))
	for _, agent := range agents {
		distros := agent.GetCapabilitiesByPrefix("kubernetes-deploy-")
		// Extract distribution names from capabilities
		distributions := make([]string, 0, len(distros))
		for _, d := range distros {
			if len(d) > len("kubernetes-deploy-") {
				distributions = append(distributions, d[len("kubernetes-deploy-"):])
			}
		}
		enrichedAgents = append(enrichedAgents, AgentWithDistros{
			Agent:                  agent,
			SupportedDistributions: distributions,
		})
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"kubernetesDeployAgents": enrichedAgents,
	})
}

func handleListKubernetesDistributions(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	// Only deployable distributions (can be installed via agents)
	distributions := []map[string]interface{}{
		{
			"id":          "K3S",
			"name":        "K3s",
			"description": "Lightweight Kubernetes by Rancher - single binary, low resource usage",
			"deployable":  true,
		},
		{
			"id":          "RKE2",
			"name":        "RKE2",
			"description": "Rancher Kubernetes Engine 2 - security-focused, FIPS compliant",
			"deployable":  true,
		},
		{
			"id":          "KUBEADM",
			"name":        "kubeadm",
			"description": "Official Kubernetes installer - standard, full-featured",
			"deployable":  true,
		},
		{
			"id":          "K0S",
			"name":        "k0s",
			"description": "Zero friction Kubernetes by Mirantis - simple, certified",
			"deployable":  true,
		},
		{
			"id":          "MICROK8S",
			"name":        "MicroK8s",
			"description": "Canonical's lightweight K8s - snap-based, addons system",
			"deployable":  true,
		},
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"kubernetesDistributions": distributions,
	})
}

func handleListAllKubernetesDistributions(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	// All distributions: deployable + connect-only
	distributions := []map[string]interface{}{
		// Deployable distributions
		{
			"id":          "K3S",
			"name":        "K3s",
			"description": "Lightweight Kubernetes by Rancher - single binary, low resource usage",
			"deployable":  true,
		},
		{
			"id":          "RKE2",
			"name":        "RKE2",
			"description": "Rancher Kubernetes Engine 2 - security-focused, FIPS compliant",
			"deployable":  true,
		},
		{
			"id":          "KUBEADM",
			"name":        "kubeadm",
			"description": "Official Kubernetes installer - standard, full-featured",
			"deployable":  true,
		},
		{
			"id":          "K0S",
			"name":        "k0s",
			"description": "Zero friction Kubernetes by Mirantis - simple, certified",
			"deployable":  true,
		},
		{
			"id":          "MICROK8S",
			"name":        "MicroK8s",
			"description": "Canonical's lightweight K8s - snap-based, addons system",
			"deployable":  true,
		},
		// Connect-only distributions (managed/existing clusters)
		{
			"id":          "EKS",
			"name":        "Amazon EKS",
			"description": "Amazon Elastic Kubernetes Service - AWS managed",
			"deployable":  false,
		},
		{
			"id":          "GKE",
			"name":        "Google GKE",
			"description": "Google Kubernetes Engine - GCP managed",
			"deployable":  false,
		},
		{
			"id":          "AKS",
			"name":        "Azure AKS",
			"description": "Azure Kubernetes Service - Azure managed",
			"deployable":  false,
		},
		{
			"id":          "OPENSHIFT",
			"name":        "OpenShift",
			"description": "Red Hat OpenShift Container Platform",
			"deployable":  false,
		},
		{
			"id":          "RANCHER",
			"name":        "Rancher",
			"description": "Rancher managed Kubernetes cluster",
			"deployable":  false,
		},
		{
			"id":          "OTHER",
			"name":        "Other",
			"description": "Other or unknown Kubernetes distribution",
			"deployable":  false,
		},
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"allKubernetesDistributions": distributions,
	})
}

func handleDeployCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input := &DeployClusterInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if distribution, ok := inputRaw["distribution"].(string); ok {
		// Validate distribution enum
		if err := graphql.ValidateEnum(distribution, graphql.KubernetesDistroValues, "distribution"); err != nil {
			graphql.WriteValidationError(w, err.Error())
			return
		}
		input.Distribution = KubernetesDistribution(distribution)
	}
	if version, ok := inputRaw["version"].(string); ok {
		input.Version = version
	}
	if masterNodes, ok := inputRaw["masterNodes"].([]interface{}); ok {
		// Limit number of nodes
		if len(masterNodes) > validation.MaxArrayLength {
			graphql.WriteValidationError(w, "too many master nodes")
			return
		}
		input.MasterNodes = make([]string, 0, len(masterNodes))
		for _, n := range masterNodes {
			if s, ok := n.(string); ok {
				input.MasterNodes = append(input.MasterNodes, s)
			}
		}
	}
	if workerNodes, ok := inputRaw["workerNodes"].([]interface{}); ok {
		// Limit number of nodes
		if len(workerNodes) > validation.MaxArrayLength {
			graphql.WriteValidationError(w, "too many worker nodes")
			return
		}
		input.WorkerNodes = make([]string, 0, len(workerNodes))
		for _, n := range workerNodes {
			if s, ok := n.(string); ok {
				input.WorkerNodes = append(input.WorkerNodes, s)
			}
		}
	}

	// Validation
	v := validation.NewValidator()
	v.Required("name", input.Name).MaxLength("name", input.Name, validation.MaxNameLength)
	if input.Description != "" {
		v.MaxLength("description", input.Description, validation.MaxDescriptionLength)
	}
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	if input.Distribution == "" {
		graphql.WriteValidationError(w, "distribution is required")
		return
	}
	if len(input.MasterNodes) == 0 {
		graphql.WriteValidationError(w, "at least one master node is required")
		return
	}

	cluster, err := service.Deploy(ctx, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "deploy cluster")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DEPLOY_CLUSTER",
		ResourceType: "cluster",
		ResourceID:   cluster.ID.String(),
		Details: map[string]interface{}{
			"name":         cluster.Name,
			"distribution": cluster.Distribution,
			"masterNodes":  len(input.MasterNodes),
			"workerNodes":  len(input.WorkerNodes),
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deployCluster": cluster,
	})
}

func handleBulkDeleteClusters(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	// Use validated bulk IDs with max limit (100)
	ids, err := graphql.ParseBulkUUIDs(variables, "ids")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	deleted, err := service.BulkDelete(ctx, tenantID, ids)
	if err != nil {
		graphql.WriteError(w, err, "bulk delete clusters")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "BULK_DELETE_CLUSTERS",
		ResourceType: "cluster",
		ResourceID:   "",
		Details: map[string]interface{}{
			"count": deleted,
			"ids":   ids,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"bulkDeleteClusters": deleted,
	})
}
