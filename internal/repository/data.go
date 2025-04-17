// Package repository provides data access implementations.
package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// contextTxKey is the context key for storing tx value
type contextTxKey struct{}

// Data manages data access resources like database connections and caches.
type Data struct {
	db  *gorm.DB
	rds *redis.Client
}

// NewData creates a new Data instance with the given resources.
func NewData(db *gorm.DB, rds *redis.Client) *Data {
	return &Data{
		db:  db,
		rds: rds,
	}
}

// Transaction executes a function within a database transaction.
// It automatically commits the transaction if the function returns without error,
// and automatically rolls back the transaction otherwise.
func (d *Data) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create a new context with transaction
		txCtx := context.WithValue(ctx, contextTxKey{}, tx)
		return fn(txCtx)
	})
}

// DB returns the database connection, honoring any transaction in the context.
func (d *Data) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return d.db.WithContext(ctx)
	}
	return tx
}
