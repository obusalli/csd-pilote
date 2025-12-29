package commands

import (
	"fmt"
	"os"

	"csd-pilote/cli/modules/commands/database"
)

// DatabaseCommand returns the database command
func DatabaseCommand() *Command {
	return &Command{
		Name:        "database",
		Aliases:     []string{"db"},
		Description: "Database management commands",
		Handler:     handleDatabase,
		SubCommands: []SubCommand{
			{Name: "init", Description: "Initialize database schema"},
			{Name: "update", Description: "Update database schema"},
			{Name: "status", Description: "Show database connection status"},
			{Name: "clean", Description: "Truncate all tables (keeps schema)"},
			{Name: "reset", Description: "Reset database (drop and recreate tables)"},
			{Name: "backup", Description: "Create database backup"},
		},
	}
}

func handleDatabase(args []string) {
	if len(args) == 0 {
		printDatabaseUsage()
		os.Exit(1)
	}

	subCommand := args[0]
	subArgs := args[1:]

	switch subCommand {
	case "init":
		database.Init(subArgs)
	case "update":
		database.Update(subArgs)
	case "status":
		database.Status(subArgs)
	case "clean":
		database.Clean(subArgs)
	case "reset":
		database.Reset(subArgs)
	case "backup":
		database.Backup(subArgs)
	case "connect":
		fmt.Println("Use: csd-corectl database connect --schema=csd_pilote")
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subCommand)
		printDatabaseUsage()
		os.Exit(1)
	}
}

func printDatabaseUsage() {
	fmt.Println("Usage: csd-pilotectl database <subcommand>")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  init        Initialize database schema")
	fmt.Println("  update      Update database schema (GORM AutoMigrate)")
	fmt.Println("  status      Show database connection status")
	fmt.Println("  clean       Truncate all tables (keeps schema)")
	fmt.Println("  reset       Reset database (drop and recreate tables)")
	fmt.Println("  backup      Create database backup")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  csd-pilotectl database init")
	fmt.Println("  csd-pilotectl db update")
	fmt.Println("  csd-pilotectl db status")
	fmt.Println("  csd-pilotectl db backup --format=sql")
	fmt.Println("  csd-pilotectl db clean --force")
	fmt.Println("  csd-pilotectl db reset --force")
	fmt.Println()
	fmt.Println("Use csd-corectl for:")
	fmt.Println("  csd-corectl database connect --schema=csd_pilote")
}
