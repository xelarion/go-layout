package types

import "mime/multipart"

// UploadFileReq 上传文件请求
type UploadFileReq struct {
	File   *multipart.FileHeader `form:"file" binding:"required"`    // 文件
	Public bool                  `form:"public" binding:"omitempty"` // 是否公开访问
}

// UploadFileResp 上传文件响应
type UploadFileResp struct {
	Key      string `json:"key"`       // 文件唯一标识
	Filename string `json:"filename"`  // 原始文件名
	Size     int64  `json:"size"`      // 文件大小（字节）
	MimeType string `json:"mime_type"` // MIME类型
	Url      string `json:"url"`       // 资源链接
}

// GetDownloadURLReq 获取下载 URL 请求
// 根据Blob Key获取下载URL，暂不校验数据权限
type GetDownloadURLReq struct {
	Key string `uri:"key" binding:"required" swaggerignore:"true"` // Blob Key
}

// GetDownloadURLResp 获取下载URL响应
type GetDownloadURLResp struct {
	URL string `json:"url"` // 下载URL（预签名）
}
