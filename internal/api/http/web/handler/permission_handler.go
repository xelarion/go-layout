package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/pkg/binding"
	"github.com/xelarion/go-layout/pkg/errs"
)

// PermissionHandler handles permission related requests
type PermissionHandler struct {
	permissionService *service.PermissionService
	logger            *zap.Logger
}

// NewPermissionHandler creates a new instance of PermissionHandler
func NewPermissionHandler(permissionService *service.PermissionService, logger *zap.Logger) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		logger:            logger.Named("web_permission_handler"),
	}
}

func (h *PermissionHandler) GetPermissionTree(c *gin.Context) {
	var req types.GetPermissionTreeReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.permissionService.GetPermissionTree(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	var req types.GetRolePermissionsReq
	if err := binding.Bind(c, &req, binding.URI); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.permissionService.GetRolePermissions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

func (h *PermissionHandler) UpdateRolePermissions(c *gin.Context) {
	var req types.UpdateRolePermissionsReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.permissionService.UpdateRolePermissions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("permissions updated successfully"))
}
