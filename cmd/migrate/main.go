package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/xelarion/go-layout/pkg/config"
	"github.com/xelarion/go-layout/pkg/migrate"
)

func main() {
	// Define flags
	migrationsDir := flag.String("dir", "migrations", "Directory with migration files")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Command is required
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	// Get command
	command := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create migrator
	m, err := migrate.NewMigrator(&cfg.PG, *migrationsDir, *verbose)
	if err != nil {
		fmt.Printf("Error creating migrator: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	// Execute command
	if err := executeCommand(m, command, args[1:]); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func executeCommand(m *migrate.Migrator, command string, args []string) error {
	switch command {
	case "up":
		return m.Up()

	case "down":
		return m.Down()

	case "reset":
		return m.Reset()

	case "status":
		return m.Status()

	case "create":
		if len(args) < 1 {
			return fmt.Errorf("create requires a migration name")
		}

		migrationType := "sql"
		if len(args) > 1 {
			migrationType = args[1]
		}

		return m.Create(args[0], migrationType)

	case "version":
		return m.Version()

	case "redo":
		return m.Redo()

	case "up-to":
		if len(args) < 1 {
			return fmt.Errorf("up-to requires a version number")
		}

		version, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid version number: %v", err)
		}

		return m.UpTo(version)

	case "down-to":
		if len(args) < 1 {
			return fmt.Errorf("down-to requires a version number")
		}

		version, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid version number: %v", err)
		}

		return m.DownTo(version)

	case "fix":
		return m.Fix()

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  up                Apply all available migrations")
	fmt.Println("  down              Rollback the last migration")
	fmt.Println("  reset             Rollback all migrations")
	fmt.Println("  status            Print migration status")
	fmt.Println("  create NAME TYPE  Create a new migration file (type: sql or go)")
	fmt.Println("  version           Print current migration version")
	fmt.Println("  redo              Rollback and reapply the latest migration")
	fmt.Println("  up-to VERSION     Migrate up to a specific version")
	fmt.Println("  down-to VERSION   Migrate down to a specific version")
	fmt.Println("  fix               Fix migrations order")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
}
