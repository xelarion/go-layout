package model

// Resource 表示系统中的资源接口，可以用于任何需要标识资源类型和ID的场景
type Resource interface {
	// GetResourceType 返回资源类型标识符
	GetResourceType() string

	// GetResourceID 返回资源唯一标识符
	GetResourceID() uint
}
