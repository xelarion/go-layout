package types

type DownloadImportTemplateReq struct {
	Type string `form:"type" binding:"required"` // 模板类型
}

type ImportCustomersReq struct {
	Files CommonUploadedFilesReq `json:"files" binding:"required,len=1"` // 上传的文件 keys
}

type ImportCustomersResp struct {
}

type ImportEquipmentTypesReq struct {
	Files CommonUploadedFilesReq `json:"files" binding:"required,len=1"` // 上传的文件 keys
}

type ImportEquipmentTypesResp struct {
}

type ImportEquipmentsReq struct {
	Files CommonUploadedFilesReq `json:"files" binding:"required,len=1"` // 上传的文件 keys
}

type ImportEquipmentsResp struct {
}

type ImportSuppliersReq struct {
	Files CommonUploadedFilesReq `json:"files" binding:"required,len=1"` // 上传的文件 keys
}

type ImportSuppliersResp struct {
}
