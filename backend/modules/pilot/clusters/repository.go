package clusters

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"csd-pilote/backend/modules/platform/database"
	"csd-pilote/backend/modules/platform/filters"
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
	if err := r.db.Create(cluster).Error; err != nil {
		return fmt.Errorf("failed to create cluster %q: %w", cluster.Name, err)
	}
	return nil
}

// GetByID retrieves a cluster by ID
func (r *Repository) GetByID(tenantID, id uuid.UUID) (*Cluster, error) {
	var cluster Cluster
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&cluster).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster %s: %w", id, err)
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
		return nil, 0, fmt.Errorf("failed to count clusters: %w", err)
	}

	// Get results with nodes preloaded (limit nodes per cluster for safety)
	if err := query.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Limit(1000).Order("role, created_at")
	}).Order("created_at DESC").Limit(limit).Offset(offset).Find(&clusters).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list clusters: %w", err)
	}

	return clusters, count, nil
}

// Update updates a cluster
func (r *Repository) Update(cluster *Cluster) error {
	if err := r.db.Save(cluster).Error; err != nil {
		return fmt.Errorf("failed to update cluster %s: %w", cluster.ID, err)
	}
	return nil
}

// Delete deletes a cluster and its associated nodes (cascade)
func (r *Repository) Delete(tenantID, id uuid.UUID) error {
	// First delete all associated nodes (cascade)
	if err := r.db.Where("cluster_id = ?", id).Delete(&ClusterNode{}).Error; err != nil {
		return fmt.Errorf("failed to delete cluster nodes for %s: %w", id, err)
	}
	// Then delete the cluster
	if err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Cluster{}).Error; err != nil {
		return fmt.Errorf("failed to delete cluster %s: %w", id, err)
	}
	return nil
}

// UpdateStatus updates the status of a cluster
func (r *Repository) UpdateStatus(tenantID, id uuid.UUID, status ClusterStatus, message string) error {
	if err := r.db.Model(&Cluster{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":          status,
			"status_message":  message,
			"last_checked_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
		return fmt.Errorf("failed to update cluster status %s: %w", id, err)
	}
	return nil
}

// CreateNodes creates multiple cluster nodes
func (r *Repository) CreateNodes(nodes []ClusterNode) error {
	if len(nodes) == 0 {
		return nil
	}
	if err := r.db.Create(&nodes).Error; err != nil {
		return fmt.Errorf("failed to create cluster nodes: %w", err)
	}
	return nil
}

// GetNodes retrieves nodes for a cluster (limited for safety)
func (r *Repository) GetNodes(clusterID uuid.UUID) ([]ClusterNode, error) {
	var nodes []ClusterNode
	if err := r.db.Where("cluster_id = ?", clusterID).Order("role, created_at").Limit(1000).Find(&nodes).Error; err != nil {
		return nil, fmt.Errorf("failed to get nodes for cluster %s: %w", clusterID, err)
	}
	return nodes, nil
}

// UpdateNodeStatus updates the status of a cluster node
func (r *Repository) UpdateNodeStatus(nodeID uuid.UUID, status, message string) error {
	if err := r.db.Model(&ClusterNode{}).
		Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"status":  status,
			"message": message,
		}).Error; err != nil {
		return fmt.Errorf("failed to update node status %s: %w", nodeID, err)
	}
	return nil
}

// DeleteNodes deletes all nodes for a cluster
func (r *Repository) DeleteNodes(clusterID uuid.UUID) error {
	if err := r.db.Where("cluster_id = ?", clusterID).Delete(&ClusterNode{}).Error; err != nil {
		return fmt.Errorf("failed to delete nodes for cluster %s: %w", clusterID, err)
	}
	return nil
}

// UpdateClusterArtifact updates the artifact key and agent ID for a deployed cluster
func (r *Repository) UpdateClusterArtifact(clusterID uuid.UUID, artifactKey string) error {
	if err := r.db.Model(&Cluster{}).
		Where("id = ?", clusterID).
		Updates(map[string]interface{}{
			"artifact_key": artifactKey,
		}).Error; err != nil {
		return fmt.Errorf("failed to update cluster artifact %s: %w", clusterID, err)
	}
	return nil
}

// GetByIDWithNodes retrieves a cluster with its nodes (limited for safety)
func (r *Repository) GetByIDWithNodes(tenantID, id uuid.UUID) (*Cluster, error) {
	var cluster Cluster
	err := r.db.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Limit(1000).Order("role, created_at")
	}).Where("tenant_id = ? AND id = ?", tenantID, id).First(&cluster).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster with nodes %s: %w", id, err)
	}
	return &cluster, nil
}

// Count returns the total count of clusters for a tenant
func (r *Repository) Count(tenantID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&Cluster{}).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count clusters: %w", err)
	}
	return count, nil
}

// CountByStatus returns the count of clusters by status for a tenant
func (r *Repository) CountByStatus(tenantID uuid.UUID, status ClusterStatus) (int64, error) {
	var count int64
	if err := r.db.Model(&Cluster{}).Where("tenant_id = ? AND status = ?", tenantID, status).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count clusters by status: %w", err)
	}
	return count, nil
}

// BulkDelete deletes multiple clusters and their associated nodes (cascade)
func (r *Repository) BulkDelete(tenantID uuid.UUID, ids []uuid.UUID) (int64, error) {
	var rowsAffected int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// First delete all associated nodes (cascade)
		if err := tx.Where("cluster_id IN ?", ids).Delete(&ClusterNode{}).Error; err != nil {
			return fmt.Errorf("failed to delete cluster nodes: %w", err)
		}
		// Then delete the clusters
		result := tx.Where("tenant_id = ? AND id IN ?", tenantID, ids).Delete(&Cluster{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete clusters: %w", result.Error)
		}
		rowsAffected = result.RowsAffected
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("bulk delete clusters failed: %w", err)
	}
	return rowsAffected, nil
}

// CountWithFilter returns the count of clusters matching the filter
func (r *Repository) CountWithFilter(tenantID uuid.UUID, filter *ClusterFilter, advancedFilter interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&Cluster{}).Where("tenant_id = ?", tenantID)

	// Apply simple filter
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

	// Apply advanced filter
	if advancedFilter != nil {
		qb := filters.NewQueryBuilder(r.db).
			WithFieldMappings(map[string]string{
				"createdAt":     "created_at",
				"updatedAt":     "updated_at",
				"statusMessage": "status_message",
				"artifactKey":   "artifact_key",
			})
		var err error
		query, err = qb.ApplyFilterJSON(query, advancedFilter)
		if err != nil {
			return 0, err
		}
	}

	err := query.Count(&count).Error
	return count, err
}
