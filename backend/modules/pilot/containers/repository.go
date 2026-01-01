package containers

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"csd-pilote/backend/modules/platform/database"
	"csd-pilote/backend/modules/platform/filters"
)

// Repository handles database operations for container engines
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new container engine repository
func NewRepository() *Repository {
	return &Repository{db: database.GetDB()}
}

// Create creates a new container engine
func (r *Repository) Create(engine *ContainerEngine) error {
	return r.db.Create(engine).Error
}

// GetByID retrieves a container engine by ID
func (r *Repository) GetByID(tenantID, id uuid.UUID) (*ContainerEngine, error) {
	var engine ContainerEngine
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&engine).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get container engine %s: %w", id, err)
	}
	return &engine, nil
}

// List retrieves all container engines for a tenant with optional filtering
func (r *Repository) List(tenantID uuid.UUID, filter *ContainerEngineFilter, limit, offset int) ([]ContainerEngine, int64, error) {
	var engines []ContainerEngine
	var count int64

	query := r.db.Model(&ContainerEngine{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.EngineType != nil {
			query = query.Where("engine_type = ?", *filter.EngineType)
		}
	}

	// Get count
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&engines).Error; err != nil {
		return nil, 0, err
	}

	return engines, count, nil
}

// Update updates a container engine
func (r *Repository) Update(engine *ContainerEngine) error {
	return r.db.Save(engine).Error
}

// Delete deletes a container engine
func (r *Repository) Delete(tenantID, id uuid.UUID) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ContainerEngine{}).Error
}

// UpdateStatus updates the status of a container engine
func (r *Repository) UpdateStatus(tenantID, id uuid.UUID, status EngineStatus, message string) error {
	return r.db.Model(&ContainerEngine{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":          status,
			"status_message":  message,
			"last_checked_at": gorm.Expr("NOW()"),
		}).Error
}

// UpdateInfo updates the cached info of a container engine
func (r *Repository) UpdateInfo(tenantID, id uuid.UUID, info map[string]interface{}) error {
	info["last_checked_at"] = gorm.Expr("NOW()")
	return r.db.Model(&ContainerEngine{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(info).Error
}

// Count returns the total count of container engines for a tenant
func (r *Repository) Count(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&ContainerEngine{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// CountByStatus returns the count of container engines by status for a tenant
func (r *Repository) CountByStatus(tenantID uuid.UUID, status EngineStatus) (int64, error) {
	var count int64
	err := r.db.Model(&ContainerEngine{}).Where("tenant_id = ? AND status = ?", tenantID, status).Count(&count).Error
	return count, err
}

// BulkDelete deletes multiple container engines by IDs
func (r *Repository) BulkDelete(tenantID uuid.UUID, ids []uuid.UUID) (int64, error) {
	result := r.db.Where("tenant_id = ? AND id IN ?", tenantID, ids).Delete(&ContainerEngine{})
	return result.RowsAffected, result.Error
}

// CountWithFilter returns the count of container engines matching the filter
func (r *Repository) CountWithFilter(tenantID uuid.UUID, filter *ContainerEngineFilter, advancedFilter interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&ContainerEngine{}).Where("tenant_id = ?", tenantID)

	// Apply simple filter
	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.EngineType != nil {
			query = query.Where("engine_type = ?", *filter.EngineType)
		}
	}

	// Apply advanced filter
	if advancedFilter != nil {
		qb := filters.NewQueryBuilder(r.db).
			WithFieldMappings(map[string]string{
				"createdAt":      "created_at",
				"updatedAt":      "updated_at",
				"statusMessage":  "status_message",
				"artifactKey":    "artifact_key",
				"engineType":     "engine_type",
				"engineVersion":  "engine_version",
				"containerCount": "container_count",
				"imageCount":     "image_count",
				"lastCheckedAt":  "last_checked_at",
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
