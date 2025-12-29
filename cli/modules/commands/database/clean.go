package database

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"csd-pilote/cli/modules/platform/backend"
	"csd-pilote/cli/modules/platform/config"
)

// Clean truncates all tables but keeps the schema
func Clean(args []string) {
	force := hasFlag(args, "--force") || hasFlag(args, "-f")

	if !force {
		fmt.Println("WARNING: This will delete all data in csd_pilote tables!")
		fmt.Print("Are you sure? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(confirm)

		if confirm != "yes" {
			fmt.Println("Aborted")
			return
		}
	}

	backend.RunWithBackend(func(cfg *config.Config) error {
		fmt.Println("Cleaning CSD-Pilote database tables...")
		fmt.Println()

		db := backend.GetDB()

		tables := []string{
			"clusters",
			"hypervisors",
			"container_engines",
		}

		for _, table := range tables {
			backend.PrintInfo("Truncating %s...", table)
			if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE csd_pilote.%s CASCADE", table)).Error; err != nil {
				backend.PrintWarning("Failed to truncate %s: %v", table, err)
			}
		}

		fmt.Println()
		backend.PrintSuccess("Database tables cleaned")

		return nil
	})
}
