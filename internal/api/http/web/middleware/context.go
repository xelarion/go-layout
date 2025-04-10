package middleware

import (
	"context"

	"github.com/xelarion/go-layout/internal/model"
)

// contextKey is a unexported type for context keys to avoid collisions
type contextKey struct{}

var (
	// CurrentKey is the key used to store current context information
	CurrentKey = &contextKey{}
)

// Current represents the current context information.
type Current struct {
	User *model.User
	// Can be extended with other fields in the future
}

// NewCurrent creates a new Current instance
func NewCurrent(user *model.User) *Current {
	return &Current{
		User: user,
	}
}

// SetCurrent stores the current context information in the context.
func SetCurrent(ctx context.Context, current *Current) context.Context {
	return context.WithValue(ctx, CurrentKey, current)
}

// GetCurrent retrieves the current context information.
// Returns nil if no current information is found.
func GetCurrent(ctx context.Context) *Current {
	if v := ctx.Value(CurrentKey); v != nil {
		if current, ok := v.(*Current); ok {
			return current
		}
	}
	return nil
}
