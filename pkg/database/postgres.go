// Package database provides database connection functionality.
package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/xelarion/go-layout/pkg/config"
)

// PostgresDB represents a PostgreSQL database connection.
type PostgresDB struct {
	DB *gorm.DB
}

// NewPostgres creates a new PostgreSQL database connection.
func NewPostgres(cfg *config.PG, log *zap.Logger) (*PostgresDB, error) {
	logLevel := logger.Silent
	if log.Core().Enabled(zap.DebugLevel) {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.New(
			&zapGormWriter{log: log.Named("postgres")},
			logger.Config{
				SlowThreshold:             time.Second, // Log SQL slower than this threshold
				LogLevel:                  logLevel,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
	}

	db, err := gorm.Open(postgres.Open(cfg.URL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection.
func (p *PostgresDB) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// zapGormWriter is a custom writer for GORM that uses zap for logging.
type zapGormWriter struct {
	log *zap.Logger
}

// Printf implements the gorm logger interface.
func (w *zapGormWriter) Printf(format string, args ...any) {
	w.log.Sugar().Debugf(format, args...)
}
