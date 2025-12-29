package database

import (
	"fmt"
	"time"

	"csd-pilote/backend/modules/platform/database/migrations"
	"csd-pilote/cli/modules/platform/backend"
	"csd-pilote/cli/modules/platform/config"
)

// Update updates the database schema
func Update(args []string) {
	verbose := hasFlag(args, "--verbose") || hasFlag(args, "-v")
	migrations.Verbose = verbose

	backend.RunWithBackend(func(cfg *config.Config) error {
		fmt.Println("Updating CSD-Pilote database schema...")
		fmt.Println()

		start := time.Now()

		// Set the backend DB for migrations
		migrations.DB = backend.GetDB()

		// Run migrations
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
		backend.PrintSuccess("Database updated successfully in %v", elapsed.Round(time.Millisecond))

		return nil
	})
}
