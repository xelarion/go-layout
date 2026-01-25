package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"

	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/pkg/binding"
	"github.com/xelarion/go-layout/pkg/errs"
)

// ImportHandler defines the import HTTP handlers.
type ImportHandler struct {
	importService *service.ImportService
	logger        *zap.Logger
}

// NewImportHandler creates a new instance of ImportHandler.
func NewImportHandler(importService *service.ImportService, logger *zap.Logger) *ImportHandler {
	return &ImportHandler{
		importService: importService,
		logger:        logger.Named("web_import_handler"),
	}
}

// ImportCustomers godoc
//
//	@ID				ImportCustomers
//	@Summary		Import Customers
//	@Description	Import Customers
//	@Tags			import
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.ImportCustomersReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.ImportCustomersResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/import/customers [post]
func (h *ImportHandler) ImportCustomers(c *gin.Context) {
	var req types.ImportCustomersReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.importService.ImportCustomers(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("操作成功，请在 系统管理-系统查询-批量操作记录 中查看结果"))
}

// ImportEquipmentTypes godoc
//
//	@ID				ImportEquipmentTypes
//	@Summary		Import Equipment Types
//	@Description	Import Equipment Types
//	@Tags			import
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.ImportEquipmentTypesReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.ImportEquipmentTypesResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/import/equipment_types [post]
func (h *ImportHandler) ImportEquipmentTypes(c *gin.Context) {
	var req types.ImportEquipmentTypesReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.importService.ImportEquipmentTypes(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("操作成功，请在 系统管理-系统查询-批量操作记录 中查看结果"))
}

// ImportEquipments godoc
//
//	@ID				ImportEquipments
//	@Summary		Import Equipments
//	@Description	Import Equipments
//	@Tags			import
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.ImportEquipmentsReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.ImportEquipmentsResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/import/equipments [post]
func (h *ImportHandler) ImportEquipments(c *gin.Context) {
	var req types.ImportEquipmentsReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.importService.ImportEquipments(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("操作成功，请在 系统管理-系统查询-批量操作记录 中查看结果"))
}

// ImportSuppliers godoc
//
//	@ID				ImportSuppliers
//	@Summary		Import Suppliers
//	@Description	Import Suppliers
//	@Tags			import
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.ImportSuppliersReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.ImportSuppliersResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/import/suppliers [post]
func (h *ImportHandler) ImportSuppliers(c *gin.Context) {
	var req types.ImportSuppliersReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.importService.ImportSuppliers(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("操作成功，请在 系统管理-系统查询-批量操作记录 中查看结果"))
}

// DownloadImportTemplate godoc
//
//	@ID				DownloadImportTemplate
//	@Summary		Download Import Template
//	@Description	Download Import Template
//	@Tags			import
//	@Accept			json
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param			req	query		types.DownloadImportTemplateReq	false	"req"
//	@Success		200	{file}		file							"Excel 文件"
//	@Failure		400	{object}	types.Response					"Bad request"
//	@Failure		401	{object}	types.Response					"Unauthorized"
//	@Failure		500	{object}	types.Response					"Internal server error"
//	@Security		BearerAuth
//	@Router			/import/template/download [get]
func (h *ImportHandler) DownloadImportTemplate(c *gin.Context) {
	var req types.DownloadImportTemplateReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	data, filename, err := h.importService.DownloadImportTemplate(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// RFC 5987 编码文件名以支持中文
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
