// Package migrate provides database migration utilities.
package migrate

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"

	"github.com/xelarion/go-layout/pkg/config"
)

// Migrator handles database migrations using goose.
type Migrator struct {
	db      *sql.DB
	dir     string
	dialect string
	verbose bool
}

// NewMigrator creates a new migrator instance.
func NewMigrator(cfg *config.PG, migrationsDir string, verbose bool) (*Migrator, error) {
	// Parse connection string and open database
	db, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Setup goose
	goose.SetBaseFS(nil) // Use native file system
	if verbose {
		goose.SetLogger(log.New(log.Writer(), "", log.LstdFlags))
	}

	return &Migrator{
		db:      db,
		dir:     migrationsDir,
		dialect: "postgres", // Default to postgres, can be extended for MySQL
		verbose: verbose,
	}, nil
}

// Close closes the database connection.
func (m *Migrator) Close() error {
	return m.db.Close()
}

// Up applies all available migrations.
func (m *Migrator) Up() error {
	// Allow out-of-order migrations for development
	return goose.Up(m.db, m.dir, goose.WithAllowMissing())
}

// Down rolls back a single migration.
func (m *Migrator) Down() error {
	return goose.Down(m.db, m.dir)
}

// Reset rolls back all migrations.
func (m *Migrator) Reset() error {
	return goose.Reset(m.db, m.dir)
}

// Status prints the migration status.
func (m *Migrator) Status() error {
	return goose.Status(m.db, m.dir)
}

// Create creates a new migration file with the given name.
func (m *Migrator) Create(name, migrationType string) error {
	// Default to timestamp versioning for development
	return goose.Create(m.db, m.dir, name, migrationType)
}

// Version prints the current migration version.
func (m *Migrator) Version() error {
	return goose.Version(m.db, m.dir)
}

// Redo rolls back and reapplies the latest migration.
func (m *Migrator) Redo() error {
	return goose.Redo(m.db, m.dir)
}

// UpTo migrates up to a specific version.
func (m *Migrator) UpTo(version int64) error {
	// Allow out-of-order migrations for development
	return goose.UpTo(m.db, m.dir, version, goose.WithAllowMissing())
}

// DownTo migrates down to a specific version.
func (m *Migrator) DownTo(version int64) error {
	return goose.DownTo(m.db, m.dir, version)
}

// Fix fixes migrations by converting timestamp versioning to sequential,
// while preserving the timestamp ordering. This is intended to be used
// before deploying to production.
func (m *Migrator) Fix() error {
	return goose.Fix(m.dir)
}
