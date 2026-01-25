package types

// TreeNodeResp represents a node in the response
type TreeNodeResp struct {
	Key      any             `json:"key"`                // key
	Title    string          `json:"title"`              // title
	Children []*TreeNodeResp `json:"children,omitempty"` // child nodes
}

// CommonUploadedFileResp 通用已上传文件响应
type CommonUploadedFileResp struct {
	Key      string `json:"key"`       // 文件唯一标识
	Filename string `json:"filename"`  // 原始文件名
	Size     int64  `json:"size"`      // 文件大小（字节）
	MimeType string `json:"mime_type"` // MIME类型
	Url      string `json:"url"`       // 资源链接
}

// CommonUploadedFilesReq 通用已上传文件请求(在具体业务中上传的数据)
type CommonUploadedFilesReq []CommonUploadedFilesReqField

type CommonUploadedFilesReqField struct {
	Key string `json:"key" binding:"required"` // 文件唯一标识
}

func (c CommonUploadedFilesReq) Keys() []string {
	keys := make([]string, len(c))
	for i, file := range c {
		keys[i] = file.Key
	}
	return keys
}
