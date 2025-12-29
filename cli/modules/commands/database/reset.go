package database

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"csd-pilote/cli/modules/platform/backend"
	"csd-pilote/cli/modules/platform/config"
)

// Reset drops and recreates the database schema
func Reset(args []string) {
	force := hasFlag(args, "--force") || hasFlag(args, "-f")

	if !force {
		fmt.Println("WARNING: This will delete all data in csd_pilote schema!")
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
		fmt.Println("Resetting CSD-Pilote database...")
		fmt.Println()

		db := backend.GetDB()

		// Drop schema
		backend.PrintInfo("Dropping schema csd_pilote...")
		if err := db.Exec("DROP SCHEMA IF EXISTS csd_pilote CASCADE").Error; err != nil {
			return fmt.Errorf("failed to drop schema: %w", err)
		}
		backend.PrintSuccess("Schema dropped")

		fmt.Println()
		backend.PrintInfo("Run 'csd-pilotectl database init' to reinitialize")

		return nil
	})
}
