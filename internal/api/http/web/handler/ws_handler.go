package handler

import (
	"net/http"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/infra/wsocket"
)

// WSHandler WebSocket 处理器
type WSHandler struct {
	manager *wsocket.Manager
	authMW  *jwt.GinJWTMiddleware
	logger  *zap.Logger
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(manager *wsocket.Manager, authMW *jwt.GinJWTMiddleware, logger *zap.Logger) *WSHandler {
	h := &WSHandler{
		manager: manager,
		authMW:  authMW,
		logger:  logger.Named("ws_handler"),
	}
	h.registerHandlers()
	return h
}

// registerHandlers 注册消息处理器
func (h *WSHandler) registerHandlers() {
	// 示例：h.manager.Handle("action_name", h.handleActionName)
}

// 示例 handler：
// func (h *WSHandler) handleActionName(ctx *wsocket.Context) {
//     var req struct { ... }
//     if err := ctx.UnmarshalData(&req); err != nil { return }
//     ctx.Reply(map[string]any{"result": "ok"})
// }

// WSConnect godoc
//
//	@ID				WSConnect
//	@Summary		WebSocket Connect
//	@Description	建立 WebSocket 连接，需要在 URL 中传递 token 参数
//	@Tags			ws
//	@Accept			json
//	@Produce		json
//	@Param			token	query	string	true	"JWT Token"
//	@Success		101		"Switching Protocols"
//	@Failure		401		{object}	types.Response	"Unauthorized"
//	@Router			/ws [get]
func (h *WSHandler) WSConnect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, types.Error(types.CodeUnauthorized, "token is required"))
		return
	}

	// 设置 Authorization header 以便 gin-jwt 解析
	c.Request.Header.Set("Authorization", "Bearer "+token)
	claims, err := h.authMW.GetClaimsFromJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.Error(types.CodeUnauthorized, "invalid token"))
		return
	}

	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, types.Error(types.CodeUnauthorized, "invalid token claims"))
		return
	}

	sourceID := strconv.Itoa(int(userIDFloat))
	if _, err = h.manager.Upgrade(c.Writer, c.Request, sourceID); err != nil {
		h.logger.Error("failed to upgrade connection", zap.Error(err))
	}
}
