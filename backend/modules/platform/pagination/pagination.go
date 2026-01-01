package pagination

import (
	"csd-pilote/backend/modules/platform/config"
)

// Params represents validated pagination parameters
type Params struct {
	Limit  int
	Offset int
}

// Normalize validates and normalizes pagination parameters using config values
// Returns normalized limit and offset values
func Normalize(limit, offset int) Params {
	cfg := config.GetConfig()

	defaultLimit := 20
	maxLimit := 100

	if cfg != nil {
		if cfg.Pagination.DefaultLimit > 0 {
			defaultLimit = cfg.Pagination.DefaultLimit
		}
		if cfg.Pagination.MaxLimit > 0 {
			maxLimit = cfg.Pagination.MaxLimit
		}
	}

	// Apply default if not specified
	if limit <= 0 {
		limit = defaultLimit
	}

	// Cap at max limit
	if limit > maxLimit {
		limit = maxLimit
	}

	// Ensure offset is non-negative
	if offset < 0 {
		offset = 0
	}

	return Params{
		Limit:  limit,
		Offset: offset,
	}
}

// NormalizeLimit validates and normalizes just the limit parameter
func NormalizeLimit(limit int) int {
	return Normalize(limit, 0).Limit
}

// DefaultLimit returns the configured default limit
func DefaultLimit() int {
	cfg := config.GetConfig()
	if cfg != nil && cfg.Pagination.DefaultLimit > 0 {
		return cfg.Pagination.DefaultLimit
	}
	return 20
}

// MaxLimit returns the configured max limit
func MaxLimit() int {
	cfg := config.GetConfig()
	if cfg != nil && cfg.Pagination.MaxLimit > 0 {
		return cfg.Pagination.MaxLimit
	}
	return 100
}
