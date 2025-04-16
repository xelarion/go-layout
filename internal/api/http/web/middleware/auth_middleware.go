package middleware

import (
	"encoding/base64"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
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
func NewAuthMiddleware(cfg *config.JWT, uc *usecase.UserUseCase, logger *zap.Logger) (*jwt.GinJWTMiddleware, error) {
	// Initialize JWT middleware with RESTful API best practices
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:                 "go-layout",
		Key:                   []byte(cfg.Secret),
		Timeout:               cfg.TokenExpiration,   // Short-lived access token
		MaxRefresh:            cfg.RefreshExpiration, // How long a user can refresh token without login
		IdentityKey:           IdentityKey,
		PayloadFunc:           payloadFunc(cfg.TokenExpiration),
		IdentityHandler:       identityHandler,
		Authenticator:         authenticator(uc),
		Authorizator:          authorizator(uc, logger),
		Unauthorized:          unauthorized,
		LoginResponse:         loginResponse,
		LogoutResponse:        logoutResponse,
		RefreshResponse:       refreshResponse,
		TokenLookup:           "header: Authorization",
		TokenHeadName:         "Bearer",
		TimeFunc:              time.Now,
		HTTPStatusMessageFunc: httpStatusMessageFunc,
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
func authenticator(uc *usecase.UserUseCase) func(c *gin.Context) (any, error) {
	return func(c *gin.Context) (any, error) {
		var req types.LoginReq
		if err := binding.Bind(c, &req, binding.JSON); err != nil {
			return nil, errs.WrapValidation(err, err.Error())
		}

		// check captcha
		if !captcha.VerifyString(req.CaptchaID, req.Captcha) {
			return nil, errs.NewBusiness("captcha is invalid")
		}

		privateKey, err := uc.GetRSAPrivateKey(c.Request.Context(), req.Key)
		if err != nil {
			return nil, err
		}

		username, err := base64.StdEncoding.DecodeString(req.Username)
		if err != nil {
			return nil, errs.WrapInternal(err, "failed to decode username")
		}
		// Decrypt username
		decryptedUsername, err := util.RSADecrypt(username, privateKey)
		if err != nil {
			return nil, err
		}

		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			return nil, errs.WrapInternal(err, "failed to decode password")
		}

		// Decrypt password
		decryptedPassword, err := util.RSADecrypt(password, privateKey)
		if err != nil {
			return nil, err
		}

		// Delete RSA private key from cache
		if err := uc.DeleteRSAPrivateKey(c.Request.Context(), req.Key); err != nil {
			return nil, err
		}

		user, err := uc.Login(c.Request.Context(), string(decryptedUsername), string(decryptedPassword))
		if err != nil {
			return nil, err
		}

		return &User{
			ID: user.ID,
		}, nil
	}
}

// authorizator determines if a user has access and stores user in context
func authorizator(uc *usecase.UserUseCase, logger *zap.Logger) func(data any, c *gin.Context) bool {
	return func(data any, c *gin.Context) bool {
		if v, ok := data.(*User); ok {
			// Get full user from database
			user, err := uc.GetByID(c.Request.Context(), v.ID)
			if err != nil {
				c.Set("jwt_error", err.Error())
				return false
			}

			if !user.Enabled {
				c.Set("jwt_error", "user account is disabled")
				return false
			}

			// Store full user in request context using Current
			c.Request = c.Request.WithContext(SetCurrent(c.Request.Context(), NewCurrent(user.ID, user.RoleID, user.RoleSlug)))

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

	if message == "user account is disabled" {
		respCode = types.CodeUserDisabled
	}

	c.JSON(code, types.Error(respCode, message))
}

// loginResponse handles login responses with the standard response format
func loginResponse(c *gin.Context, code int, token string, expire time.Time) {
	// Calculate seconds until expiration
	expiresIn := int64(expire.Sub(time.Now()).Seconds())

	c.JSON(code, types.Success(types.LoginResp{
		Token:     token,
		Expire:    types.Time(expire),
		ExpiresIn: expiresIn,
		TokenType: TokenType,
	}))
}

// logoutResponse handles logout responses with the standard response format
func logoutResponse(c *gin.Context, code int) {
	c.JSON(code, types.Success(types.LogoutResp{}))
}

// refreshResponse handles refresh token responses with the standard response format
func refreshResponse(c *gin.Context, code int, token string, expire time.Time) {
	// Calculate seconds until expiration
	expiresIn := int64(expire.Sub(time.Now()).Seconds())

	c.JSON(code, types.Success(types.RefreshTokenResp{
		Token:     token,
		Expire:    types.Time(expire),
		ExpiresIn: expiresIn,
		TokenType: TokenType,
	}))
}

func httpStatusMessageFunc(err error, c *gin.Context) string {
	if msg, exists := c.Get("jwt_error"); exists {
		return msg.(string)
	}
	return err.Error()
}
