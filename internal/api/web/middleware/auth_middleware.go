// Package middleware provides HTTP middleware components specifically for Web API.
package middleware

import (
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/binding"
	"github.com/xelarion/go-layout/pkg/config"
	"github.com/xelarion/go-layout/pkg/errs"
)

const (
	IdentityKey = "id"
)

// User represents the user identity in JWT claims
type User struct {
	ID uint `json:"id"`
}

// JWTConfig holds JWT middleware configuration
type JWTConfig struct {
	Secret         string
	Expiration     time.Duration
	MaxRefresh     time.Duration
	TokenLookup    string
	TokenHeadName  string
	SendCookie     bool
	SecureCookie   bool
	CookieHTTPOnly bool
	CookieDomain   string
	CookieName     string
	CookieSameSite http.SameSite
}

// DefaultJWTConfig returns a production-ready JWT configuration
func DefaultJWTConfig(cfg *config.JWT) *JWTConfig {
	return &JWTConfig{
		Secret:         cfg.Secret,
		Expiration:     cfg.Expiration,
		MaxRefresh:     time.Hour * 24 * 7, // 7 days
		TokenLookup:    "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:  "Bearer",
		SendCookie:     true,
		SecureCookie:   true, // Enable in production
		CookieHTTPOnly: true,
		CookieDomain:   "", // Set according to actual deployment
		CookieName:     "jwt",
		CookieSameSite: http.SameSiteNoneMode,
	}
}

// NewAuthMiddleware creates a new JWT auth middleware with production-ready configuration.
func NewAuthMiddleware(cfg *config.JWT, userUseCase *usecase.UserUseCase, logger *zap.Logger) (*jwt.GinJWTMiddleware, error) {
	jwtConfig := DefaultJWTConfig(cfg)

	// Initialize JWT middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "go-layout",
		Key:             []byte(jwtConfig.Secret),
		Timeout:         jwtConfig.Expiration,
		MaxRefresh:      jwtConfig.MaxRefresh,
		IdentityKey:     IdentityKey,
		PayloadFunc:     payloadFunc,
		IdentityHandler: identityHandler,
		Authenticator:   authenticator(userUseCase),
		Authorizator:    authorizator(userUseCase),
		Unauthorized:    unauthorized,
		LoginResponse:   loginResponse,
		TokenLookup:     jwtConfig.TokenLookup,
		TokenHeadName:   jwtConfig.TokenHeadName,
		TimeFunc:        time.Now,
		SendCookie:      jwtConfig.SendCookie,
		SecureCookie:    jwtConfig.SecureCookie,
		CookieHTTPOnly:  jwtConfig.CookieHTTPOnly,
		CookieDomain:    jwtConfig.CookieDomain,
		CookieName:      jwtConfig.CookieName,
		CookieSameSite:  jwtConfig.CookieSameSite,
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
			"exp":       time.Now().Add(time.Minute * 30).Unix(), // Token expiration
			"iat":       time.Now().Unix(),                       // Token issued at
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
func loginResponse(c *gin.Context, code int, token string, time time.Time) {
	c.JSON(code, types.Success(types.LoginResp{
		Token:  token,
		Expire: time,
	}))
}
