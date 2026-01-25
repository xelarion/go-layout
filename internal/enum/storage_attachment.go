package enum

// StorageAttachmentNameType StorageAttachment 附件类型名称, 如: file, logo (区分同一条记录不同类型的附件)
type StorageAttachmentNameType string

const (
	StorageAttachmentNameTypeNone    StorageAttachmentNameType = ""
	StorageAttachmentNameTypeFile    StorageAttachmentNameType = "file"
	StorageAttachmentNameTypeImage   StorageAttachmentNameType = "image"
	StorageAttachmentNameTypeContent StorageAttachmentNameType = "content" // 富文本内容中的文件
)
