package clusters

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
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
	if agentId, ok := inputRaw["agentId"].(string); ok {
		input.AgentID = agentId
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}
	if distribution, ok := inputRaw["distribution"].(string); ok {
		input.Distribution = KubernetesDistribution(distribution)
	}

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}
	if input.AgentID == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("agentId is required"))
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
	if agentId, ok := inputRaw["agentId"].(string); ok {
		input.AgentID = agentId
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}
	if distribution, ok := inputRaw["distribution"].(string); ok {
		input.Distribution = KubernetesDistribution(distribution)
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

func handleListKubernetesAgents(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	token, _ := middleware.GetTokenFromContext(ctx)

	client := csdcore.GetClient()
	agents, err := client.ListAgentsByCapability(ctx, token, "kubernetes")
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"kubernetesAgents": agents,
	}))
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
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"kubernetesDeployAgents": enrichedAgents,
	}))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"kubernetesDistributions": distributions,
	}))
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

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"allKubernetesDistributions": distributions,
	}))
}

func handleDeployCluster(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
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

	input := &DeployClusterInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if distribution, ok := inputRaw["distribution"].(string); ok {
		input.Distribution = KubernetesDistribution(distribution)
	}
	if version, ok := inputRaw["version"].(string); ok {
		input.Version = version
	}
	if masterNodes, ok := inputRaw["masterNodes"].([]interface{}); ok {
		input.MasterNodes = make([]string, 0, len(masterNodes))
		for _, n := range masterNodes {
			if s, ok := n.(string); ok {
				input.MasterNodes = append(input.MasterNodes, s)
			}
		}
	}
	if workerNodes, ok := inputRaw["workerNodes"].([]interface{}); ok {
		input.WorkerNodes = make([]string, 0, len(workerNodes))
		for _, n := range workerNodes {
			if s, ok := n.(string); ok {
				input.WorkerNodes = append(input.WorkerNodes, s)
			}
		}
	}

	// Validation
	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}
	if input.Distribution == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("distribution is required"))
		return
	}
	if len(input.MasterNodes) == 0 {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("at least one master node is required"))
		return
	}

	cluster, err := service.Deploy(ctx, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deployCluster": cluster,
	}))
}
