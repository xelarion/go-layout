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

// DepartmentHandler defines the department HTTP handlers.
type DepartmentHandler struct {
	departmentService *service.DepartmentService
	logger            *zap.Logger
}

// NewDepartmentHandler creates a new instance of DepartmentHandler.
func NewDepartmentHandler(departmentService *service.DepartmentService, logger *zap.Logger) *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: departmentService,
		logger:            logger.Named("web_department_handler"),
	}
}

// CreateDepartment godoc
//
//	@ID				CreateDepartment
//	@Summary		Create Department
//	@Description	Creates a new Department
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.CreateDepartmentReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.CreateDepartmentResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/departments [post]
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req types.CreateDepartmentReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.departmentService.CreateDepartment(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("Created successfully"))
}

// ListDepartments godoc
//
//	@ID				ListDepartments
//	@Summary		List Departments
//	@Description	Retrieves a list of Departments
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListDepartmentsReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.ListDepartmentsResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/departments [get]
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	var req types.ListDepartmentsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.departmentService.ListDepartments(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetDepartment godoc
//
//	@ID				GetDepartment
//	@Summary		Get Department
//	@Description	Retrieves a single Department
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	query		types.GetDepartmentReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetDepartmentResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/departments/{id} [get]
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	var req types.GetDepartmentReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.departmentService.GetDepartment(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, types.Error(types.CodeNotFound, "Department not found"))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetDepartmentFormData godoc
//
//	@ID				GetDepartmentFormData
//	@Summary		Get Department Form Data
//	@Description	Retrieves a single Department Form Data
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer													true	"id"
//	@Param			req	query		types.GetDepartmentFormDataReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetDepartmentFormDataResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/departments/{id}/form [get]
func (h *DepartmentHandler) GetDepartmentFormData(c *gin.Context) {
	var req types.GetDepartmentFormDataReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.departmentService.GetDepartmentFormData(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateDepartment godoc
//
//	@ID				UpdateDepartment
//	@Summary		Update Department
//	@Description	Updates an existing Department
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	body		types.UpdateDepartmentReq						true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateDepartmentResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/departments/{id} [put]
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	var req types.UpdateDepartmentReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.departmentService.UpdateDepartment(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("Updated successfully"))
}

// UpdateDepartmentEnabled godoc
//
//	@ID				UpdateDepartmentEnabled
//	@Summary		Update Department Enabled
//	@Description	Updates an existing Department Enabled
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer													true	"id"
//	@Param			req	body		types.UpdateDepartmentEnabledReq						true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateDepartmentEnabledResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/departments/{id}/enabled [patch]
func (h *DepartmentHandler) UpdateDepartmentEnabled(c *gin.Context) {
	var req types.UpdateDepartmentEnabledReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.departmentService.UpdateDepartmentEnabled(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("Operation successfully"))
}

// DeleteDepartment godoc
//
//	@ID				DeleteDepartment
//	@Summary		Delete Department
//	@Description	Deletes an existing Department
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	body		types.DeleteDepartmentReq						true	"req"
//	@Success		204	{object}	types.Response{data=types.DeleteDepartmentResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/departments/{id} [delete]
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	var req types.DeleteDepartmentReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.departmentService.DeleteDepartment(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("Deleted successfully"))
}
