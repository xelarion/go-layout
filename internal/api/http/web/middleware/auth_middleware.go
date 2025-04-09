// Package middleware provides HTTP middleware components specifically for Web API.
package middleware

import (
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/binding"
	"github.com/xelarion/go-layout/pkg/config"
	"github.com/xelarion/go-layout/pkg/errs"
)

const (
	IdentityKey = "id"
	TokenType   = "Bearer" // Standard token type for JWT
)

// User represents the user identity in JWT claims
type User struct {
	ID uint `json:"id"`
}

// NewAuthMiddleware creates a new JWT auth middleware with production-ready configuration.
func NewAuthMiddleware(cfg *config.JWT, userUseCase *usecase.UserUseCase, logger *zap.Logger) (*jwt.GinJWTMiddleware, error) {
	// Initialize JWT middleware with RESTful API best practices
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "go-layout",
		Key:             []byte(cfg.Secret),
		Timeout:         cfg.TokenExpiration,   // Short-lived access token
		MaxRefresh:      cfg.RefreshExpiration, // How long a user can refresh token without login
		IdentityKey:     IdentityKey,
		PayloadFunc:     payloadFunc(cfg.TokenExpiration),
		IdentityHandler: identityHandler,
		Authenticator:   authenticator(userUseCase),
		Authorizator:    authorizator(userUseCase),
		Unauthorized:    unauthorized,
		LoginResponse:   loginResponse,
		RefreshResponse: refreshResponse,
		// REST API best practice: use header for token transport
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		// Disable cookie for REST API (SPA frontend)
		SendCookie: false,
	})

	if err != nil {
		logger.Error("Failed to initialize JWT middleware", zap.Error(err))
		return nil, err
	}

	return authMiddleware, nil
}

// payloadFunc is used to set the JWT payload
func payloadFunc(tokenExpiration time.Duration) func(data any) jwt.MapClaims {
	return func(data any) jwt.MapClaims {
		if v, ok := data.(*User); ok {
			now := time.Now()
			return jwt.MapClaims{
				IdentityKey: v.ID,
				"exp":       now.Add(tokenExpiration).Unix(), // Token expiration
				"iat":       now.Unix(),                      // Token issued at
				"nbf":       now.Unix(),                      // Not valid before
			}
		}
		return jwt.MapClaims{}
	}
}

// identityHandler sets the identity for the JWT claims
func identityHandler(c *gin.Context) any {
	claims := jwt.ExtractClaims(c)
	return &User{
		ID: uint(claims[IdentityKey].(float64)),
	}
}

// authenticator validates user credentials and returns identity
func authenticator(userUseCase *usecase.UserUseCase) func(c *gin.Context) (any, error) {
	return func(c *gin.Context) (any, error) {
		var req types.LoginReq
		if err := binding.Bind(c, &req, binding.JSON); err != nil {
			return nil, errs.WrapValidation(err, err.Error())
		}

		user, err := userUseCase.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			return nil, err
		}

		return &User{
			ID: user.ID,
		}, nil
	}
}

// authorizator determines if a user has access and stores user in context
func authorizator(userUseCase *usecase.UserUseCase) func(data any, c *gin.Context) bool {
	return func(data any, c *gin.Context) bool {
		if v, ok := data.(*User); ok {
			// Get full user from database
			user, err := userUseCase.GetByID(c.Request.Context(), v.ID)
			if err != nil {
				_ = c.Error(err)
				return false
			}

			// Store full user in request context using Current
			c.Request = c.Request.WithContext(SetCurrent(c.Request.Context(), NewCurrent(user)))

			return true
		}
		return false
	}
}

// unauthorized handles unauthorized responses with the standard response format
func unauthorized(c *gin.Context, code int, message string) {
	var respCode int
	switch code {
	case http.StatusUnauthorized:
		respCode = types.CodeUnauthorized
	case http.StatusForbidden:
		respCode = types.CodeForbidden
	default:
		respCode = types.CodeUnauthorized
	}

	c.JSON(code, types.Error(respCode, message))
}

// loginResponse handles login responses with the standard response format
func loginResponse(c *gin.Context, code int, token string, expire time.Time) {
	// Calculate seconds until expiration
	expiresIn := int64(expire.Sub(time.Now()).Seconds())

	c.JSON(code, types.Success(types.LoginResp{
		Token:     token,
		Expire:    expire,
		ExpiresIn: expiresIn,
		TokenType: TokenType,
	}))
}

// refreshResponse handles refresh token responses with the standard response format
func refreshResponse(c *gin.Context, code int, token string, expire time.Time) {
	// Calculate seconds until expiration
	expiresIn := int64(expire.Sub(time.Now()).Seconds())

	c.JSON(code, types.Success(types.RefreshResp{
		Token:     token,
		Expire:    expire,
		ExpiresIn: expiresIn,
		TokenType: TokenType,
	}))
}
