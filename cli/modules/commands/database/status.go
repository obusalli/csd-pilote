package database

import (
	"fmt"

	"csd-pilote/cli/modules/platform/backend"
	"csd-pilote/cli/modules/platform/config"
)

// Status shows database connection status
func Status(args []string) {
	backend.RunWithBackend(func(cfg *config.Config) error {
		fmt.Println("CSD-Pilote Database Status")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()

		db := backend.GetDB()

		// Check if schema exists
		var schemaExists bool
		db.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'csd_pilote')").Scan(&schemaExists)

		if schemaExists {
			backend.PrintSuccess("Schema: csd_pilote exists")
			fmt.Println()

			// Count records in each table
			var clusterCount, hypervisorCount, containerEngineCount int64

			db.Raw("SELECT COUNT(*) FROM csd_pilote.clusters").Scan(&clusterCount)
			db.Raw("SELECT COUNT(*) FROM csd_pilote.hypervisors").Scan(&hypervisorCount)
			db.Raw("SELECT COUNT(*) FROM csd_pilote.container_engines").Scan(&containerEngineCount)

			fmt.Println("Tables:")
			backend.PrintInfo("Clusters: %d", clusterCount)
			backend.PrintInfo("Hypervisors: %d", hypervisorCount)
			backend.PrintInfo("Container Engines: %d", containerEngineCount)
		} else {
			backend.PrintWarning("Schema: csd_pilote does not exist")
			fmt.Println()
			backend.PrintInfo("Run 'csd-pilotectl database init' to initialize")
		}

		fmt.Println()
		return nil
	})
}
