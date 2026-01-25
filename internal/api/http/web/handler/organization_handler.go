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

// OrganizationHandler defines the organization HTTP handlers.
type OrganizationHandler struct {
	organizationService *service.OrganizationService
	logger              *zap.Logger
}

// NewOrganizationHandler creates a new instance of OrganizationHandler.
func NewOrganizationHandler(organizationService *service.OrganizationService, logger *zap.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		organizationService: organizationService,
		logger:              logger.Named("web_organization_handler"),
	}
}

// CreateOrganization godoc
//
//	@ID				CreateOrganization
//	@Summary		Create Organization
//	@Description	Creates a new Organization
//	@Tags			organization
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.CreateOrganizationReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.CreateOrganizationResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req types.CreateOrganizationReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.organizationService.CreateOrganization(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("created successfully"))
}

// ListOrganizations godoc
//
//	@ID				ListOrganizations
//	@Summary		List Organizations
//	@Description	Retrieves a list of Organizations
//	@Tags			organization
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListOrganizationsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListOrganizationsResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/organizations [get]
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	var req types.ListOrganizationsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.organizationService.ListOrganizations(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetOrganizationFormData godoc
//
//	@ID				GetOrganizationFormData
//	@Summary		Get Organization Form Data
//	@Description	Retrieves a single Organization Form Data
//	@Tags			organization
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer													true	"id"
//	@Param			req	query		types.GetOrganizationFormDataReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetOrganizationFormDataResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/organizations/{id}/form [get]
func (h *OrganizationHandler) GetOrganizationFormData(c *gin.Context) {
	var req types.GetOrganizationFormDataReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.organizationService.GetOrganizationFormData(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateOrganization godoc
//
//	@ID				UpdateOrganization
//	@Summary		Update Organization
//	@Description	Updates an existing Organization
//	@Tags			organization
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	body		types.UpdateOrganizationReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateOrganizationResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	var req types.UpdateOrganizationReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.organizationService.UpdateOrganization(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("updated successfully"))
}

// DeleteOrganization godoc
//
//	@ID				DeleteOrganization
//	@Summary		Delete Organization
//	@Description	Deletes an existing Organization
//	@Tags			organization
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	body		types.DeleteOrganizationReq							true	"req"
//	@Success		204	{object}	types.Response{data=types.DeleteOrganizationResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	var req types.DeleteOrganizationReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.organizationService.DeleteOrganization(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("deleted successfully"))
}

// GetOrganizationsTree godoc
//
//	@ID				GetOrganizationsTree
//	@Summary		Get Organization Tree
//	@Description	Retrieves a single Organization Tree
//	@Tags			organization
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetOrganizationsTreeReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetOrganizationsTreeResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/organizations/tree [get]
func (h *OrganizationHandler) GetOrganizationsTree(c *gin.Context) {
	var req types.GetOrganizationsTreeReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.organizationService.GetOrganizationsTree(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
