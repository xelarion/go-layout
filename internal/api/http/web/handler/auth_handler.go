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

// NewCaptcha godoc
//
//	@ID				NewCaptcha
//	@Summary		New Captcha
//	@Description	New Captcha
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.NewCaptchaReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.NewCaptchaResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Router			/captcha/new [post]
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

// ReloadCaptcha godoc
//
//	@ID				ReloadCaptcha
//	@Summary		Reload Captcha
//	@Description	Reload Captcha
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string											true	"id"
//	@Param			req	body		types.ReloadCaptchaReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.ReloadCaptchaResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Router			/captcha/{id}/reload [post]
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

// GetRSAPublicKey godoc
//
//	@ID				GetRSAPublicKey
//	@Summary		Get RSAPublic Key
//	@Description	Retrieves a single RSAPublic Key
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.GetRSAPublicKeyReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.GetRSAPublicKeyResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Router			/public_key [post]
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

// Login godoc
//
//	@ID				Login
//	@Summary		Login
//	@Description	Login
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.LoginReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.LoginResp}	"Success"
//	@Failure		400	{object}	types.Response							"Bad request"
//	@Failure		401	{object}	types.Response							"Unauthorized"
//	@Failure		500	{object}	types.Response							"Internal server error"
//	@Router			/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	h.authMW.LoginHandler(c)
}

// RefreshToken godoc
//
//	@ID				RefreshToken
//	@Summary		Refresh Token
//	@Description	Refresh Token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.RefreshTokenReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.RefreshTokenResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Router			/refresh_token [get]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	h.authMW.RefreshHandler(c)
}

// Logout godoc
//
//	@ID				Logout
//	@Summary		Logout
//	@Description	Logout
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.LogoutReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.LogoutResp}	"Success"
//	@Failure		400	{object}	types.Response							"Bad request"
//	@Failure		401	{object}	types.Response							"Unauthorized"
//	@Failure		500	{object}	types.Response							"Internal server error"
//	@Security		BearerAuth
//	@Router			/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	h.authMW.LogoutHandler(c)
}

// GetProfile godoc
//
//	@ID				GetProfile
//	@Summary		Get Profile
//	@Description	Retrieves a single Profile
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetProfileReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetProfileResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/profile [get]
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

// UpdateProfile godoc
//
//	@ID				UpdateProfile
//	@Summary		Update Profile
//	@Description	Updates an existing Profile
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.UpdateProfileReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateProfileResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/profile [put]
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

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("Operation successfully"))
}

// GetCurrentUserInfo godoc
//
//	@ID				GetCurrentUserInfo
//	@Summary		Get Current User Info
//	@Description	Retrieves a single Current User Info
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetCurrentUserInfoReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetCurrentUserInfoResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/current [get]
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
