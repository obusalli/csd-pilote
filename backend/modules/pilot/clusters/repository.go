package clusters

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"csd-pilote/backend/modules/platform/database"
)

// Repository handles database operations for clusters
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new cluster repository
func NewRepository() *Repository {
	return &Repository{db: database.GetDB()}
}

// Create creates a new cluster
func (r *Repository) Create(cluster *Cluster) error {
	return r.db.Create(cluster).Error
}

// GetByID retrieves a cluster by ID
func (r *Repository) GetByID(tenantID, id uuid.UUID) (*Cluster, error) {
	var cluster Cluster
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// List retrieves all clusters for a tenant with optional filtering
func (r *Repository) List(tenantID uuid.UUID, filter *ClusterFilter, limit, offset int) ([]Cluster, int64, error) {
	var clusters []Cluster
	var count int64

	query := r.db.Model(&Cluster{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Mode != nil {
			query = query.Where("mode = ?", *filter.Mode)
		}
		if filter.Distribution != nil {
			query = query.Where("distribution = ?", *filter.Distribution)
		}
	}

	// Get count
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get results with nodes preloaded
	if err := query.Preload("Nodes").Order("created_at DESC").Limit(limit).Offset(offset).Find(&clusters).Error; err != nil {
		return nil, 0, err
	}

	return clusters, count, nil
}

// Update updates a cluster
func (r *Repository) Update(cluster *Cluster) error {
	return r.db.Save(cluster).Error
}

// Delete deletes a cluster
func (r *Repository) Delete(tenantID, id uuid.UUID) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Cluster{}).Error
}

// UpdateStatus updates the status of a cluster
func (r *Repository) UpdateStatus(tenantID, id uuid.UUID, status ClusterStatus, message string) error {
	return r.db.Model(&Cluster{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":          status,
			"status_message":  message,
			"last_checked_at": gorm.Expr("NOW()"),
		}).Error
}

// CreateNodes creates multiple cluster nodes
func (r *Repository) CreateNodes(nodes []ClusterNode) error {
	if len(nodes) == 0 {
		return nil
	}
	return r.db.Create(&nodes).Error
}

// GetNodes retrieves all nodes for a cluster
func (r *Repository) GetNodes(clusterID uuid.UUID) ([]ClusterNode, error) {
	var nodes []ClusterNode
	err := r.db.Where("cluster_id = ?", clusterID).Order("role, created_at").Find(&nodes).Error
	return nodes, err
}

// UpdateNodeStatus updates the status of a cluster node
func (r *Repository) UpdateNodeStatus(nodeID uuid.UUID, status, message string) error {
	return r.db.Model(&ClusterNode{}).
		Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"status":  status,
			"message": message,
		}).Error
}

// DeleteNodes deletes all nodes for a cluster
func (r *Repository) DeleteNodes(clusterID uuid.UUID) error {
	return r.db.Where("cluster_id = ?", clusterID).Delete(&ClusterNode{}).Error
}

// UpdateClusterArtifact updates the artifact key and agent ID for a deployed cluster
func (r *Repository) UpdateClusterArtifact(clusterID uuid.UUID, artifactKey string) error {
	return r.db.Model(&Cluster{}).
		Where("id = ?", clusterID).
		Updates(map[string]interface{}{
			"artifact_key": artifactKey,
		}).Error
}

// GetByIDWithNodes retrieves a cluster with its nodes
func (r *Repository) GetByIDWithNodes(tenantID, id uuid.UUID) (*Cluster, error) {
	var cluster Cluster
	err := r.db.Preload("Nodes").Where("tenant_id = ? AND id = ?", tenantID, id).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}
