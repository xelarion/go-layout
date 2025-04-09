package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/types"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger.Named("web_auth_handler"),
	}
}

// GetCaptcha godoc
//	@Summary		Get Captcha
//	@Description	Retrieves a single Captcha
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetCaptchaReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetCaptchaResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Router			/captcha [get]
func (h *AuthHandler) GetCaptcha(c *gin.Context) {
	// In a real application, you would use a captcha library to generate the image
	// For this example, we'll use a mock implementation
	resp := types.GetCaptchaResp{
		CaptchaID:  "sample-captcha-id",
		CaptchaImg: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAGQAAAAoCAMAAAA...", // Mock base64 image
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
