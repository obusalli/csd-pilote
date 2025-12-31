package dashboard

import (
	"context"
	"encoding/json"
	"net/http"

	"csd-pilote/backend/modules/pilot/clusters"
	"csd-pilote/backend/modules/pilot/containers"
	"csd-pilote/backend/modules/pilot/hypervisors"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
)

func init() {
	graphql.RegisterQuery("dashboardStats", "Get dashboard statistics", "csd-pilote.dashboard.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDashboardStats(ctx, w, variables)
		})
}

// DashboardStats represents the dashboard statistics
type DashboardStats struct {
	ClustersCount         int64 `json:"clustersCount"`
	ClustersConnected     int64 `json:"clustersConnected"`
	ClustersDeploying     int64 `json:"clustersDeploying"`
	ClustersError         int64 `json:"clustersError"`
	HypervisorsCount      int64 `json:"hypervisorsCount"`
	HypervisorsConnected  int64 `json:"hypervisorsConnected"`
	HypervisorsDeploying  int64 `json:"hypervisorsDeploying"`
	HypervisorsError      int64 `json:"hypervisorsError"`
	ContainerEnginesCount int64 `json:"containerEnginesCount"`
	EnginesConnected      int64 `json:"enginesConnected"`
	EnginesError          int64 `json:"enginesError"`
}

func handleDashboardStats(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	// Get cluster stats
	clusterRepo := clusters.NewRepository()
	clustersCount, _ := clusterRepo.Count(tenantID)
	clustersConnected, _ := clusterRepo.CountByStatus(tenantID, clusters.ClusterStatusConnected)
	clustersDeploying, _ := clusterRepo.CountByStatus(tenantID, clusters.ClusterStatusDeploying)
	clustersError, _ := clusterRepo.CountByStatus(tenantID, clusters.ClusterStatusError)

	// Get hypervisor stats
	hypervisorRepo := hypervisors.NewRepository()
	hypervisorsCount, _ := hypervisorRepo.Count(tenantID)
	hypervisorsConnected, _ := hypervisorRepo.CountByStatus(tenantID, hypervisors.HypervisorStatusConnected)
	hypervisorsDeploying, _ := hypervisorRepo.CountByStatus(tenantID, hypervisors.HypervisorStatusDeploying)
	hypervisorsError, _ := hypervisorRepo.CountByStatus(tenantID, hypervisors.HypervisorStatusError)

	// Get container engine stats
	containerRepo := containers.NewRepository()
	enginesCount, _ := containerRepo.Count(tenantID)
	enginesConnected, _ := containerRepo.CountByStatus(tenantID, containers.EngineStatusConnected)
	enginesError, _ := containerRepo.CountByStatus(tenantID, containers.EngineStatusError)

	stats := DashboardStats{
		ClustersCount:         clustersCount,
		ClustersConnected:     clustersConnected,
		ClustersDeploying:     clustersDeploying,
		ClustersError:         clustersError,
		HypervisorsCount:      hypervisorsCount,
		HypervisorsConnected:  hypervisorsConnected,
		HypervisorsDeploying:  hypervisorsDeploying,
		HypervisorsError:      hypervisorsError,
		ContainerEnginesCount: enginesCount,
		EnginesConnected:      enginesConnected,
		EnginesError:          enginesError,
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"dashboardStats": stats,
	}))
}
