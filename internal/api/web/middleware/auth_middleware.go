// Package middleware provides HTTP middleware components specifically for Web API.
package middleware

import (
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/config"
)

// Auth keys in context
const (
	IdentityKey = "id"
	UserKey     = "user"
)

// User represents the user identity in JWT claims
type User struct {
	ID uint `json:"id"`
}

// NewAuthMiddleware creates a new JWT auth middleware.
func NewAuthMiddleware(cfg *config.JWT, userUseCase *usecase.UserUseCase, logger *zap.Logger) (*jwt.GinJWTMiddleware, error) {
	// Initialize JWT middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "go-layout",
		Key:             []byte(cfg.Secret),
		Timeout:         cfg.Expiration,
		MaxRefresh:      time.Hour * 24 * 7, // 7 days
		IdentityKey:     IdentityKey,
		PayloadFunc:     payloadFunc,
		IdentityHandler: identityHandler,
		Authenticator:   authenticator(userUseCase, logger),
		Authorizator:    authorizator,
		Unauthorized:    unauthorized,
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenHeadName is a string in the header
		// Default value is "Bearer"
		TokenHeadName: "Bearer",
		// TimeFunc provides the current time. You can override it to use another time value.
		// This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
		// Best configuration for front-end and back-end separation projects
		SendCookie:     true,                  // Enable cookie transport
		SecureCookie:   false,                 // Set to false in development, true in production (requires HTTPS)
		CookieHTTPOnly: true,                  // Prevent XSS attacks
		CookieDomain:   "",                    // Set according to actual deployment
		CookieName:     "jwt",                 // Cookie name
		CookieSameSite: http.SameSiteNoneMode, // Same-site cookie policy, suitable for front-end and back-end separation
	})

	if err != nil {
		logger.Error("Failed to initialize JWT middleware", zap.Error(err))
		return nil, err
	}

	return authMiddleware, nil
}

// payloadFunc is used to set the JWT payload
func payloadFunc(data any) jwt.MapClaims {
	if v, ok := data.(*User); ok {
		return jwt.MapClaims{
			IdentityKey: v.ID,
		}
	}
	return jwt.MapClaims{}
}

// identityHandler sets the identity for the JWT claims
func identityHandler(c *gin.Context) any {
	claims := jwt.ExtractClaims(c)
	return &User{
		ID: uint(claims[IdentityKey].(float64)),
	}
}

// authenticator validates user credentials and returns identity
func authenticator(userUseCase *usecase.UserUseCase, logger *zap.Logger) func(c *gin.Context) (any, error) {
	return func(c *gin.Context) (any, error) {
		var loginVals struct {
			Email    string `form:"email" json:"email" binding:"required,email"`
			Password string `form:"password" json:"password" binding:"required"`
		}

		if err := c.ShouldBind(&loginVals); err != nil {
			return nil, jwt.ErrMissingLoginValues
		}

		// Get user from database and verify
		user, err := userUseCase.Login(c.Request.Context(), loginVals.Email, loginVals.Password)
		if err != nil {
			logger.Warn("Authentication failed",
				zap.String("email", loginVals.Email),
				zap.Error(err))
			return nil, jwt.ErrFailedAuthentication
		}

		return &User{
			ID: user.ID,
		}, nil
	}
}

// authorizator determines if a user has access
func authorizator(data any, c *gin.Context) bool {
	// Store user in context for later use
	if v, ok := data.(*User); ok {
		c.Set(UserKey, v)
		return true
	}
	return false
}

// unauthorized handles unauthorized responses
func unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

// GetCurrentUser extracts the current user from the context
func GetCurrentUser(c *gin.Context) *model.User {
	if v, exists := c.Get(UserKey); exists {
		if user, ok := v.(*User); ok {
			// Create a model.User from our User
			// Note: This no longer contains name and role information, complete user info needs to be fetched from database
			return &model.User{
				ID: user.ID,
			}
		}
	}
	return nil
}

// GetCurrentFromContext extracts user information and populates the context.Current structure
func GetCurrentFromContext(c *gin.Context, userUseCase *usecase.UserUseCase) *Current {
	// First check if we already have a Current in the context
	if current := GetCurrent(c.Request.Context()); current != nil {
		return current
	}

	// If not, try to get JWT user and populate from there
	if user := GetCurrentUser(c); user != nil {
		// Fetch complete user info from database
		fullUser, err := userUseCase.GetByID(c.Request.Context(), user.ID)
		// TODO
		if err != nil || fullUser == nil {
			return nil
		}

		// Create and store Current in context with user information
		current := NewCurrent(fullUser)

		// Store in request context for future use
		c.Request = c.Request.WithContext(SetCurrent(c.Request.Context(), current))

		return current
	}

	return nil
}

// AdminOnly returns a middleware that checks if the user has admin role
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the current user from context
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Unauthorized",
			})
			return
		}

		// Check if the user has admin role
		if user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "Forbidden: Admin access required",
			})
			return
		}

		c.Next()
	}
}

// AdminOnlyWithContext returns a middleware that checks if the user has admin role using the context.Current
func AdminOnlyWithContext(userUseCase *usecase.UserUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the current context
		current := GetCurrentFromContext(c, userUseCase)
		if current == nil || current.User == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Unauthorized",
			})
			return
		}

		// Check if the user has admin role
		if current.User.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "Forbidden: Admin access required",
			})
			return
		}

		c.Next()
	}
}

// AdminAuthorizatorMiddleware returns a middleware that checks if the user has admin role
// Due to design changes, this middleware no longer extracts role information from JWT,
// but suggests to judge through business logic
func AdminAuthorizatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(401, gin.H{
				"code":    401,
				"message": "Authentication required",
			})
			return
		}

		// In actual projects, user role information should be obtained from the database
		// This provides a simple example framework, which needs to be replaced with actual logic when used
		isAdmin := false // Example: Need to query user role from database

		if !isAdmin {
			c.AbortWithStatusJSON(403, gin.H{
				"code":    403,
				"message": "Admin access required",
			})
			return
		}

		c.Next()
	}
}
