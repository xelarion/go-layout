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

// StorageHandler 文件存储相关请求处理器
type StorageHandler struct {
	storageService *service.StorageService
	logger         *zap.Logger
}

// NewStorageHandler 创建StorageHandler实例
func NewStorageHandler(storageService *service.StorageService, logger *zap.Logger) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
		logger:         logger.Named("web_storage_handler"),
	}
}

// UploadFile godoc
//
//	@ID				UploadFile
//	@Summary		Upload File
//	@Description	Upload File
//	@Tags			storage
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file										true	"File to upload"
//	@Param			public	formData	bool										false	"public"
//	@Success		201		{object}	types.Response{data=types.UploadFileResp}	"Success"
//	@Failure		400		{object}	types.Response								"Bad request"
//	@Failure		401		{object}	types.Response								"Unauthorized"
//	@Failure		500		{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/storage/upload [post]
func (h *StorageHandler) UploadFile(c *gin.Context) {
	var req types.UploadFileReq
	if err := binding.Bind(c, &req, binding.FormMultipart); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.storageService.Upload(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetDownloadURL godoc
//
//	@ID				GetDownloadURL
//	@Summary		Get Download URL
//	@Description	Retrieves a single Download URL
//	@Tags			storage
//	@Accept			json
//	@Produce		json
//	@Param			key	path		string											true	"key"
//	@Param			req	query		types.GetDownloadURLReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetDownloadURLResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/storage/{key}/download_url [get]
func (h *StorageHandler) GetDownloadURL(c *gin.Context) {
	var req types.GetDownloadURLReq
	if err := binding.Bind(c, &req, binding.URI); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.storageService.GetDownloadURL(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
