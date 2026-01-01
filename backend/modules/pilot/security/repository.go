package security

import (
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"csd-pilote/backend/modules/platform/database"
	"csd-pilote/backend/modules/platform/filters"
)

// Repository handles database operations for security entities
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new security repository
func NewRepository() *Repository {
	return &Repository{db: database.GetDB()}
}

// ========================================
// Firewall Rules
// ========================================

// CreateRule creates a new firewall rule
func (r *Repository) CreateRule(rule *FirewallRule) error {
	return r.db.Create(rule).Error
}

// GetRuleByID retrieves a rule by ID
func (r *Repository) GetRuleByID(tenantID, id uuid.UUID) (*FirewallRule, error) {
	var rule FirewallRule
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// ListRules retrieves all rules for a tenant with optional filtering
func (r *Repository) ListRules(tenantID uuid.UUID, filter *FirewallRuleFilter, limit, offset int) ([]FirewallRule, int64, error) {
	var rules []FirewallRule
	var count int64

	query := r.db.Model(&FirewallRule{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ? OR comment ILIKE ?", search, search, search)
		}
		if filter.Chain != nil {
			query = query.Where("chain = ?", *filter.Chain)
		}
		if filter.Protocol != nil {
			query = query.Where("protocol = ?", *filter.Protocol)
		}
		if filter.Action != nil {
			query = query.Where("action = ?", *filter.Action)
		}
		if filter.Enabled != nil {
			query = query.Where("enabled = ?", *filter.Enabled)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("priority ASC, created_at DESC").Limit(limit).Offset(offset).Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, count, nil
}

// UpdateRule updates a firewall rule
func (r *Repository) UpdateRule(rule *FirewallRule) error {
	return r.db.Save(rule).Error
}

// DeleteRule deletes a firewall rule
func (r *Repository) DeleteRule(tenantID, id uuid.UUID) error {
	// First remove from any profiles
	r.db.Where("rule_id = ?", id).Delete(&FirewallProfileRule{})
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&FirewallRule{}).Error
}

// BulkDeleteRules deletes multiple rules by IDs
func (r *Repository) BulkDeleteRules(tenantID uuid.UUID, ids []uuid.UUID) (int64, error) {
	var rowsAffected int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// First remove from any profiles
		if err := tx.Where("rule_id IN ?", ids).Delete(&FirewallProfileRule{}).Error; err != nil {
			return err
		}
		// Then delete the rules
		result := tx.Where("tenant_id = ? AND id IN ?", tenantID, ids).Delete(&FirewallRule{})
		if result.Error != nil {
			return result.Error
		}
		rowsAffected = result.RowsAffected
		return nil
	})

	return rowsAffected, err
}

// CountRules returns the total count of rules for a tenant
func (r *Repository) CountRules(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&FirewallRule{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// CountRulesWithFilter returns the count of rules matching the filter
func (r *Repository) CountRulesWithFilter(tenantID uuid.UUID, filter *FirewallRuleFilter, advancedFilter interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&FirewallRule{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ? OR comment ILIKE ?", search, search, search)
		}
		if filter.Chain != nil {
			query = query.Where("chain = ?", *filter.Chain)
		}
		if filter.Protocol != nil {
			query = query.Where("protocol = ?", *filter.Protocol)
		}
		if filter.Action != nil {
			query = query.Where("action = ?", *filter.Action)
		}
		if filter.Enabled != nil {
			query = query.Where("enabled = ?", *filter.Enabled)
		}
	}

	if advancedFilter != nil {
		qb := filters.NewQueryBuilder(r.db).
			WithFieldMappings(map[string]string{
				"createdAt":  "created_at",
				"updatedAt":  "updated_at",
				"sourceIp":   "source_ip",
				"sourcePort": "source_port",
				"destIp":     "dest_ip",
				"destPort":   "dest_port",
				"ruleExpr":   "rule_expr",
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

// ========================================
// Firewall Profiles
// ========================================

// CreateProfile creates a new firewall profile
func (r *Repository) CreateProfile(profile *FirewallProfile) error {
	return r.db.Create(profile).Error
}

// GetProfileByID retrieves a profile by ID
func (r *Repository) GetProfileByID(tenantID, id uuid.UUID) (*FirewallProfile, error) {
	var profile FirewallProfile
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetProfileByIDWithRules retrieves a profile with its rules
func (r *Repository) GetProfileByIDWithRules(tenantID, id uuid.UUID) (*FirewallProfile, error) {
	var profile FirewallProfile
	err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("priority ASC")
	}).Where("tenant_id = ? AND id = ?", tenantID, id).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListProfiles retrieves all profiles for a tenant with optional filtering
func (r *Repository) ListProfiles(tenantID uuid.UUID, filter *FirewallProfileFilter, limit, offset int) ([]FirewallProfile, int64, error) {
	var profiles []FirewallProfile
	var count int64

	query := r.db.Model(&FirewallProfile{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
		}
		if filter.IsDefault != nil {
			query = query.Where("is_default = ?", *filter.IsDefault)
		}
		if filter.Enabled != nil {
			query = query.Where("enabled = ?", *filter.Enabled)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Rules").Order("created_at DESC").Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
		return nil, 0, err
	}

	return profiles, count, nil
}

// UpdateProfile updates a firewall profile
func (r *Repository) UpdateProfile(profile *FirewallProfile) error {
	return r.db.Save(profile).Error
}

// DeleteProfile deletes a firewall profile
func (r *Repository) DeleteProfile(tenantID, id uuid.UUID) error {
	// First remove all rule associations
	r.db.Where("profile_id = ?", id).Delete(&FirewallProfileRule{})
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&FirewallProfile{}).Error
}

// CountProfiles returns the total count of profiles for a tenant
func (r *Repository) CountProfiles(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&FirewallProfile{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// AddRulesToProfile adds rules to a profile (validates tenant ownership)
func (r *Repository) AddRulesToProfile(tenantID, profileID uuid.UUID, ruleIDs []uuid.UUID) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	// Verify all rules belong to the same tenant (security: tenant isolation)
	var count int64
	r.db.Model(&FirewallRule{}).
		Where("tenant_id = ? AND id IN ?", tenantID, ruleIDs).
		Count(&count)
	if count != int64(len(ruleIDs)) {
		return gorm.ErrRecordNotFound // Some rules don't belong to this tenant
	}

	// Get current max sort order
	var maxOrder int
	r.db.Model(&FirewallProfileRule{}).
		Where("profile_id = ?", profileID).
		Select("COALESCE(MAX(sort_order), -1)").
		Scan(&maxOrder)

	// Create associations
	for i, ruleID := range ruleIDs {
		assoc := FirewallProfileRule{
			ProfileID: profileID,
			RuleID:    ruleID,
			SortOrder: maxOrder + i + 1,
		}
		if err := r.db.Create(&assoc).Error; err != nil {
			// Ignore duplicate key errors
			continue
		}
	}
	return nil
}

// RemoveRulesFromProfile removes rules from a profile
func (r *Repository) RemoveRulesFromProfile(profileID uuid.UUID, ruleIDs []uuid.UUID) error {
	return r.db.Where("profile_id = ? AND rule_id IN ?", profileID, ruleIDs).Delete(&FirewallProfileRule{}).Error
}

// SetProfileRules replaces all rules in a profile (validates tenant ownership)
func (r *Repository) SetProfileRules(tenantID, profileID uuid.UUID, ruleIDs []uuid.UUID) error {
	// Remove all existing associations
	if err := r.db.Where("profile_id = ?", profileID).Delete(&FirewallProfileRule{}).Error; err != nil {
		return err
	}
	// Add new associations (with tenant validation)
	if len(ruleIDs) > 0 {
		return r.AddRulesToProfile(tenantID, profileID, ruleIDs)
	}
	return nil
}

// ========================================
// Firewall Templates
// ========================================

// CreateTemplate creates a new firewall template
func (r *Repository) CreateTemplate(template *FirewallTemplate) error {
	return r.db.Create(template).Error
}

// GetTemplateByID retrieves a template by ID
func (r *Repository) GetTemplateByID(tenantID, id uuid.UUID) (*FirewallTemplate, error) {
	var template FirewallTemplate
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// ListTemplates retrieves all templates for a tenant with optional filtering
func (r *Repository) ListTemplates(tenantID uuid.UUID, filter *FirewallTemplateFilter, limit, offset int) ([]FirewallTemplate, int64, error) {
	var templates []FirewallTemplate
	var count int64

	// Include system-wide templates (tenant_id is null) and tenant-specific
	query := r.db.Model(&FirewallTemplate{}).Where("tenant_id = ? OR is_built_in = true", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
		}
		if filter.Category != nil {
			query = query.Where("category = ?", *filter.Category)
		}
		if filter.IsBuiltIn != nil {
			query = query.Where("is_built_in = ?", *filter.IsBuiltIn)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("is_built_in DESC, created_at DESC").Limit(limit).Offset(offset).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, count, nil
}

// UpdateTemplate updates a firewall template
func (r *Repository) UpdateTemplate(template *FirewallTemplate) error {
	return r.db.Save(template).Error
}

// DeleteTemplate deletes a firewall template
func (r *Repository) DeleteTemplate(tenantID, id uuid.UUID) error {
	// Don't allow deleting built-in templates
	return r.db.Where("tenant_id = ? AND id = ? AND is_built_in = false", tenantID, id).Delete(&FirewallTemplate{}).Error
}

// CountTemplates returns the total count of templates for a tenant
func (r *Repository) CountTemplates(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&FirewallTemplate{}).Where("tenant_id = ? OR is_built_in = true", tenantID).Count(&count).Error
	return count, err
}

// GetTemplateRules parses and returns the rules from a template
func (r *Repository) GetTemplateRules(template *FirewallTemplate) ([]TemplateRuleDefinition, error) {
	var rules []TemplateRuleDefinition
	if template.RulesJSON == "" {
		return rules, nil
	}
	err := json.Unmarshal([]byte(template.RulesJSON), &rules)
	return rules, err
}

// ========================================
// Firewall Deployments
// ========================================

// CreateDeployment creates a new firewall deployment
func (r *Repository) CreateDeployment(deployment *FirewallDeployment) error {
	return r.db.Create(deployment).Error
}

// GetDeploymentByID retrieves a deployment by ID
func (r *Repository) GetDeploymentByID(tenantID, id uuid.UUID) (*FirewallDeployment, error) {
	var deployment FirewallDeployment
	err := r.db.Preload("Profile").Where("tenant_id = ? AND id = ?", tenantID, id).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// ListDeployments retrieves all deployments for a tenant with optional filtering
func (r *Repository) ListDeployments(tenantID uuid.UUID, filter *FirewallDeploymentFilter, limit, offset int) ([]FirewallDeployment, int64, error) {
	var deployments []FirewallDeployment
	var count int64

	query := r.db.Model(&FirewallDeployment{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("agent_name ILIKE ? OR status_message ILIKE ?", search, search)
		}
		if filter.ProfileID != nil {
			if profileID, err := uuid.Parse(*filter.ProfileID); err == nil {
				query = query.Where("profile_id = ?", profileID)
			}
		}
		if filter.AgentID != nil {
			if agentID, err := uuid.Parse(*filter.AgentID); err == nil {
				query = query.Where("agent_id = ?", agentID)
			}
		}
		if filter.Action != nil {
			query = query.Where("action = ?", *filter.Action)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Profile").Order("created_at DESC").Limit(limit).Offset(offset).Find(&deployments).Error; err != nil {
		return nil, 0, err
	}

	return deployments, count, nil
}

// UpdateDeployment updates a deployment
func (r *Repository) UpdateDeployment(deployment *FirewallDeployment) error {
	return r.db.Save(deployment).Error
}

// UpdateDeploymentStatus updates the status of a deployment
func (r *Repository) UpdateDeploymentStatus(id uuid.UUID, status DeploymentStatus, message, output string) error {
	updates := map[string]interface{}{
		"status":         status,
		"status_message": message,
	}
	if output != "" {
		updates["output"] = output
	}
	if status == DeploymentStatusDeploying {
		updates["started_at"] = gorm.Expr("NOW()")
	}
	if status == DeploymentStatusApplied || status == DeploymentStatusError || status == DeploymentStatusRolledBack {
		updates["completed_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&FirewallDeployment{}).Where("id = ?", id).Updates(updates).Error
}

// CountDeployments returns the total count of deployments for a tenant
func (r *Repository) CountDeployments(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&FirewallDeployment{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// CountDeploymentsByStatus returns the count of deployments by status
func (r *Repository) CountDeploymentsByStatus(tenantID uuid.UUID, status DeploymentStatus) (int64, error) {
	var count int64
	err := r.db.Model(&FirewallDeployment{}).Where("tenant_id = ? AND status = ?", tenantID, status).Count(&count).Error
	return count, err
}

// GetLatestDeploymentForAgent retrieves the most recent deployment for an agent
func (r *Repository) GetLatestDeploymentForAgent(tenantID, agentID uuid.UUID) (*FirewallDeployment, error) {
	var deployment FirewallDeployment
	err := r.db.Preload("Profile").
		Where("tenant_id = ? AND agent_id = ?", tenantID, agentID).
		Order("created_at DESC").
		First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}
