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

// RoleHandler defines the role HTTP handlers.
type RoleHandler struct {
	roleService *service.RoleService
	logger      *zap.Logger
}

// NewRoleHandler creates a new instance of RoleHandler.
func NewRoleHandler(roleService *service.RoleService, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger.Named("web_role_handler"),
	}
}

// CreateRole godoc
//
//	@ID				CreateRole
//	@Summary		Create Role
//	@Description	Creates a new Role
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.CreateRoleReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.CreateRoleResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req types.CreateRoleReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("created successfully"))
}

// ListRoles godoc
//
//	@ID				ListRoles
//	@Summary		List Roles
//	@Description	Retrieves a list of Roles
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListRolesReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListRolesResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	var req types.ListRolesReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.ListRoles(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetRole godoc
//
//	@ID				GetRole
//	@Summary		Get Role
//	@Description	Retrieves a single Role
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer									true	"id"
//	@Param			req	query		types.GetRoleReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetRoleResp}	"Success"
//	@Failure		400	{object}	types.Response							"Bad request"
//	@Failure		401	{object}	types.Response							"Unauthorized"
//	@Failure		500	{object}	types.Response							"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	var req types.GetRoleReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.GetRole(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetRoleFormData godoc
//
//	@ID				GetRoleFormData
//	@Summary		Get Role Form Data
//	@Description	Retrieves a single Role Form Data
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	query		types.GetRoleFormDataReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetRoleFormDataResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles/{id}/form [get]
func (h *RoleHandler) GetRoleFormData(c *gin.Context) {
	var req types.GetRoleFormDataReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.GetRoleFormData(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateRole godoc
//
//	@ID				UpdateRole
//	@Summary		Update Role
//	@Description	Updates an existing Role
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer										true	"id"
//	@Param			req	body		types.UpdateRoleReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateRoleResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	var req types.UpdateRoleReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.UpdateRole(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("updated successfully"))
}

// UpdateRoleEnabled godoc
//
//	@ID				UpdateRoleEnabled
//	@Summary		Update Role Enabled
//	@Description	Updates an existing Role Enabled
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	body		types.UpdateRoleEnabledReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateRoleEnabledResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles/{id}/enabled [patch]
func (h *RoleHandler) UpdateRoleEnabled(c *gin.Context) {
	var req types.UpdateRoleEnabledReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.UpdateRoleEnabled(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("operation successfully"))
}

// DeleteRole godoc
//
//	@ID				DeleteRole
//	@Summary		Delete Role
//	@Description	Deletes an existing Role
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer										true	"id"
//	@Param			req	body		types.DeleteRoleReq							true	"req"
//	@Success		204	{object}	types.Response{data=types.DeleteRoleResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	var req types.DeleteRoleReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.DeleteRole(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("deleted successfully"))
}

// GetRoleOptions godoc
//
//	@ID				GetRoleOptions
//	@Summary		Get Role Options
//	@Description	Retrieves a single Role Options
//	@Tags			role
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetRoleOptionsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetRoleOptionsResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/roles/options [get]
func (h *RoleHandler) GetRoleOptions(c *gin.Context) {
	var req types.GetRoleOptionsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.roleService.GetRoleOptions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
