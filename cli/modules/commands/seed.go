package commands

import (
	"fmt"
	"os"
)

// SeedCommand returns the seed command
func SeedCommand() *Command {
	return &Command{
		Name:        "seed",
		Aliases:     []string{},
		Description: "Seed data into databases",
		Handler:     handleSeed,
		SubCommands: []SubCommand{
			{Name: "all", Description: "Seed all data (use csd-corectl seed all)"},
			{Name: "reference", Description: "Seed reference data (use csd-corectl)"},
		},
	}
}

func handleSeed(args []string) {
	if len(args) == 0 {
		printSeedUsage()
		os.Exit(1)
	}

	subCommand := args[0]

	switch subCommand {
	case "all", "reference", "permissions", "menus", "templates":
		fmt.Println("Seeds are handled by csd-corectl.")
		fmt.Println()
		fmt.Println("Run from cli/ directory:")
		fmt.Println("  csd-corectl seed all")
		fmt.Println()
		fmt.Println("This will seed:")
		fmt.Println("  - Permission categories")
		fmt.Println("  - Permissions")
		fmt.Println("  - Menus")
		fmt.Println("  - Artifact types")
		fmt.Println("  - Role templates")
		fmt.Println()
		fmt.Println("Seed files location: data/seeds/core/")
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subCommand)
		printSeedUsage()
		os.Exit(1)
	}
}

func printSeedUsage() {
	fmt.Println("Usage: csd-pilotectl seed <subcommand>")
	fmt.Println()
	fmt.Println("All seeds for csd-pilote are handled by csd-corectl.")
	fmt.Println()
	fmt.Println("Run from cli/ directory:")
	fmt.Println("  csd-corectl seed all")
	fmt.Println()
	fmt.Println("Seed files are located in data/seeds/:")
	fmt.Println("  data/seeds/core/       - Core data (permissions, menus, templates)")
	fmt.Println("  data/seeds/i18n/       - Translations")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cd cli && csd-corectl seed all")
	fmt.Println("  cd cli && csd-corectl seed permissions")
	fmt.Println("  cd cli && csd-corectl seed menus")
	fmt.Println("  cd cli && csd-corectl seed templates")
}
