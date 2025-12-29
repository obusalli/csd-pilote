package database

import (
	"fmt"
	"strings"
	"time"

	"csd-pilote/backend/modules/platform/database/migrations"
	"csd-pilote/cli/modules/platform/backend"
	"csd-pilote/cli/modules/platform/config"
)

// Init initializes the database schema
func Init(args []string) {
	verbose := hasFlag(args, "--verbose") || hasFlag(args, "-v")
	migrations.Verbose = verbose

	backend.RunWithBackend(func(cfg *config.Config) error {
		fmt.Println("Initializing CSD-Pilote database...")
		fmt.Println()

		start := time.Now()

		// Set the backend DB for migrations
		migrations.DB = backend.GetDB()

		// Run migrations
		fmt.Println("Step 1: Running GORM AutoMigrate...")
		result, err := migrations.AutoMigrateWithResult()
		if err != nil {
			return fmt.Errorf("failed to auto-migrate: %w", err)
		}

		// Display results
		if result != nil {
			if result.TotalCreated > 0 || result.TotalUpdated > 0 {
				backend.PrintSuccess("Tables: %d created, %d updated, %d unchanged",
					result.TotalCreated, result.TotalUpdated, result.TotalNoChange)
			} else {
				backend.PrintInfo("All %d tables already up to date", result.TotalTables)
			}
			if result.IndexesCreated > 0 {
				backend.PrintSuccess("Indexes: %d created", result.IndexesCreated)
			}
		}

		elapsed := time.Since(start)
		fmt.Println()
		backend.PrintSuccess("Database initialized successfully in %v", elapsed.Round(time.Millisecond))

		return nil
	})
}

// hasFlag checks if a flag is present in the arguments
func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

// getFlagValue retrieves the value of a flag from the arguments
func getFlagValue(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, flag+"=") {
			return strings.TrimPrefix(arg, flag+"=")
		}
	}
	return ""
}
