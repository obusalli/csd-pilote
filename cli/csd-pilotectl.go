package main

import (
	"fmt"
	"os"

	"csd-pilote/cli/modules/commands"
)

var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "version", "-v", "--version":
		fmt.Printf("csd-pilotectl version %s\n", Version)
		return
	case "help", "-h", "--help":
		printUsage()
		return
	}

	// Get command from registry
	cmd := commands.GetCommand(command)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Run 'csd-pilotectl help' for usage.\n")
		os.Exit(1)
	}

	// Execute command
	cmd.Handler(args)
}

func printUsage() {
	fmt.Println("csd-pilotectl - CSD-Pilote CLI Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  csd-pilotectl <command> [subcommand] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  database, db    Database management (init, update, status, reset, clean, backup)")
	fmt.Println("  seed            Seed data (redirects to csd-corectl)")
	fmt.Println("  version         Show version")
	fmt.Println("  help            Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  csd-pilotectl database init       Initialize database schema")
	fmt.Println("  csd-pilotectl db update           Update database schema")
	fmt.Println("  csd-pilotectl db status           Show database status")
	fmt.Println()
	fmt.Println("Use csd-corectl for seeds:")
	fmt.Println("  csd-corectl seed all              (from cli/ directory)")
	fmt.Println("  csd-corectl database connect --schema=csd_pilote")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --config <path>    Path to configuration file")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Default config locations:")
	fmt.Println("    ./csd-pilote.yaml")
	fmt.Println("    ../backend/csd-pilote.yaml")
	fmt.Println("    /etc/csd-pilote/csd-pilote.yaml")
}
