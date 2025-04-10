package handler

import (
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/pkg/binding"
	"github.com/xelarion/go-layout/pkg/errs"
)

type AuthHandler struct {
	authService *service.AuthService
	authMW      *jwt.GinJWTMiddleware
	logger      *zap.Logger
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authService *service.AuthService, authMW *jwt.GinJWTMiddleware, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		authMW:      authMW,
		logger:      logger.Named("web_auth_handler"),
	}
}

func (h *AuthHandler) NewCaptcha(c *gin.Context) {
	var req types.NewCaptchaReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.authService.NewCaptcha(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// ReloadCaptcha generates and returns a captcha image.
func (h *AuthHandler) ReloadCaptcha(c *gin.Context) {
	var req types.ReloadCaptchaReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.authService.ReloadCaptcha(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

func (h *AuthHandler) GetRSAPublicKey(c *gin.Context) {
	var req types.GetRSAPublicKeyReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.authService.GetRSAPublicKey(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

func (h *AuthHandler) Login(c *gin.Context) {
	h.authMW.LoginHandler(c)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	h.authMW.RefreshHandler(c)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.authMW.LogoutHandler(c)
}

// GetProfile handles requests to get the current user's profile.
func (h *AuthHandler) GetProfile(c *gin.Context) {
	var req types.GetProfileReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.authService.GetProfile(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateProfile handles requests to update the current user's profile.
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req types.UpdateProfileReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.authService.UpdateProfile(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("Profile updated successfully"))
}

func (h *AuthHandler) GetCurrentUserInfo(c *gin.Context) {
	var req types.GetCurrentUserInfoReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.authService.GetCurrentUserInfo(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
