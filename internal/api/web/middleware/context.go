// Package middleware provides HTTP middleware components specifically for Web API.
package middleware

import (
	"context"

	"github.com/xelarion/go-layout/internal/model"
)

type contextKey string

const (
	// CurrentKey is the key used to store current context information.
	CurrentKey contextKey = "current_ctx"
)

// Current represents the current context information.
type Current struct {
	User *model.User
}

// NewCurrent creates a new Current instance.
func NewCurrent(user *model.User) *Current {
	return &Current{
		User: user,
	}
}

// SetCurrent stores the current context information.
func SetCurrent(ctx context.Context, current *Current) context.Context {
	return context.WithValue(ctx, CurrentKey, current)
}

// GetCurrent retrieves the current context information.
// Returns nil if no current information is found.
func GetCurrent(ctx context.Context) *Current {
	current, ok := ctx.Value(CurrentKey).(*Current)
	if !ok {
		return nil
	}
	return current
}

// WithCurrent creates a context with current information.
func WithCurrent(ctx context.Context, user *model.User) context.Context {
	return SetCurrent(ctx, NewCurrent(user))
}
