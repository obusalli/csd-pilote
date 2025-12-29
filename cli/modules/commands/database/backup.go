package database

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"csd-pilote/cli/modules/platform/backend"
	"csd-pilote/cli/modules/platform/config"
)

// Backup creates a database backup
func Backup(args []string) {
	format := getFlagValue(args, "--format")
	if format == "" {
		format = "custom" // pg_dump custom format by default
	}

	outputFile := getFlagValue(args, "--output")
	if outputFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		switch format {
		case "sql", "plain":
			outputFile = fmt.Sprintf("csd_pilote_backup_%s.sql", timestamp)
		default:
			outputFile = fmt.Sprintf("csd_pilote_backup_%s.dump", timestamp)
		}
	}

	backend.RunWithBackend(func(cfg *config.Config) error {
		fmt.Println("Creating CSD-Pilote database backup...")
		fmt.Println()

		// Parse database URL to get connection details
		dbURL := cfg.Database.URL

		var pgDumpFormat string
		switch format {
		case "sql", "plain":
			pgDumpFormat = "plain"
		case "custom":
			pgDumpFormat = "custom"
		case "tar":
			pgDumpFormat = "tar"
		case "directory":
			pgDumpFormat = "directory"
		default:
			pgDumpFormat = "custom"
		}

		// Build pg_dump command
		cmd := exec.Command("pg_dump",
			"-d", dbURL,
			"-n", "csd_pilote",
			"-F", pgDumpFormat,
			"-f", outputFile,
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		backend.PrintInfo("Running pg_dump...")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pg_dump failed: %w", err)
		}

		fmt.Println()
		backend.PrintSuccess("Backup created: %s", outputFile)

		return nil
	})
}
