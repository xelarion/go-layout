// Package storage provides storage abstraction for different storage backends.
package storage

import (
	"context"
	"io"
	"strings"
	"time"
)

// PublicPrefix 公开文件的存储路径前缀
const PublicPrefix = "public/"

// Storage 定义存储接口，抽象不同的存储后端（MinIO、OSS等）
type Storage interface {
	// PutObject 上传对象
	PutObject(ctx context.Context, key string, reader io.Reader, size int64, mimeType string, opts ...PutOption) error

	// GetURL 获取资源访问URL
	// 公开对象返回永久URL，私有对象返回预签名URL
	GetURL(ctx context.Context, key string, opts ...URLOption) (string, error)

	// Exists 检查对象是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// DeleteObject 删除对象
	DeleteObject(ctx context.Context, key string) error

	// StatObject 获取对象信息
	StatObject(ctx context.Context, key string) (*ObjectInfo, error)
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string
	Size         int64
	MimeType     string
	LastModified time.Time
	ETag         string
}

// PutOptions 上传选项
type PutOptions struct {
	Public bool // 公开访问（存储到 public/ 前缀）
}

// PutOption 上传选项函数
type PutOption func(*PutOptions)

// WithPutPublic 设置为公开访问
func WithPutPublic() PutOption {
	return func(o *PutOptions) {
		o.Public = true
	}
}

// ApplyPutOptions 应用上传选项
func ApplyPutOptions(opts ...PutOption) *PutOptions {
	o := &PutOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// URLOptions URL生成选项
type URLOptions struct {
	Download  bool   // 强制下载模式（attachment）
	Filename  string // 下载文件名
	Watermark string // 水印参数（OSS格式，如 "image/watermark,text_xxx"）
}

// URLOption URL选项函数
type URLOption func(*URLOptions)

// WithURLDownload 设置为下载模式
func WithURLDownload(filename string) URLOption {
	return func(o *URLOptions) {
		o.Download = true
		o.Filename = filename
	}
}

// WithURLWatermark 设置水印参数（OSS/COS支持，MinIO不支持）
func WithURLWatermark(params string) URLOption {
	return func(o *URLOptions) {
		o.Watermark = params
	}
}

// ApplyURLOptions 应用URL选项
func ApplyURLOptions(opts ...URLOption) *URLOptions {
	o := &URLOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// IsPublicKey 判断key是否为公开对象
func IsPublicKey(key string) bool {
	return strings.HasPrefix(key, PublicPrefix)
}
