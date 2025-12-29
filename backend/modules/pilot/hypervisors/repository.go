package hypervisors

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"csd-pilote/backend/modules/platform/database"
)

// Repository handles database operations for hypervisors
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new hypervisor repository
func NewRepository() *Repository {
	return &Repository{db: database.GetDB()}
}

// Create creates a new hypervisor
func (r *Repository) Create(hypervisor *Hypervisor) error {
	return r.db.Create(hypervisor).Error
}

// GetByID retrieves a hypervisor by ID
func (r *Repository) GetByID(tenantID, id uuid.UUID) (*Hypervisor, error) {
	var hypervisor Hypervisor
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&hypervisor).Error
	if err != nil {
		return nil, err
	}
	return &hypervisor, nil
}

// List retrieves all hypervisors for a tenant with optional filtering
func (r *Repository) List(tenantID uuid.UUID, filter *HypervisorFilter, limit, offset int) ([]Hypervisor, int64, error) {
	var hypervisors []Hypervisor
	var count int64

	query := r.db.Model(&Hypervisor{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ? OR hostname ILIKE ?", search, search, search)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
	}

	// Get count
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&hypervisors).Error; err != nil {
		return nil, 0, err
	}

	return hypervisors, count, nil
}

// Update updates a hypervisor
func (r *Repository) Update(hypervisor *Hypervisor) error {
	return r.db.Save(hypervisor).Error
}

// Delete deletes a hypervisor
func (r *Repository) Delete(tenantID, id uuid.UUID) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Hypervisor{}).Error
}

// UpdateStatus updates the status of a hypervisor
func (r *Repository) UpdateStatus(tenantID, id uuid.UUID, status HypervisorStatus, message string) error {
	return r.db.Model(&Hypervisor{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":          status,
			"status_message":  message,
			"last_checked_at": gorm.Expr("NOW()"),
		}).Error
}

// UpdateInfo updates the cached info of a hypervisor
func (r *Repository) UpdateInfo(tenantID, id uuid.UUID, info map[string]interface{}) error {
	info["last_checked_at"] = gorm.Expr("NOW()")
	return r.db.Model(&Hypervisor{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(info).Error
}
