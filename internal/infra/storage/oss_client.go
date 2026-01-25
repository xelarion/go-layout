package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/pkg/errs"
)

// OSSClient 阿里云OSS存储客户端，实现Storage接口
type OSSClient struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	urlExpires time.Duration
	publicURL  string // 公开访问地址（CDN）
}

// NewOSSClient 创建阿里云OSS客户端
func NewOSSClient(cfg *config.Storage) (*OSSClient, error) {
	if cfg.OSSEndpoint == "" {
		return nil, fmt.Errorf("oss endpoint is required")
	}

	client, err := oss.New(cfg.OSSEndpoint, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create oss client: %w", err)
	}

	o := &OSSClient{
		client:     client,
		bucketName: cfg.Bucket,
		urlExpires: cfg.URLExpires,
		publicURL:  cfg.PublicURL,
	}

	// 确保 bucket 存在
	if err := o.ensureBucket(); err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	o.bucket = bucket

	// 设置 public/ 前缀的公开读策略
	if err := o.ensurePublicPolicy(); err != nil {
		return nil, fmt.Errorf("failed to set public policy: %w", err)
	}

	return o, nil
}

// ensureBucket 确保 bucket 存在
func (o *OSSClient) ensureBucket() error {
	exists, err := o.client.IsBucketExist(o.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if exists {
		return nil
	}

	if err := o.client.CreateBucket(o.bucketName); err != nil {
		return fmt.Errorf("bucket %q create failed: %w, or you can create it manually in Aliyun console", o.bucketName, err)
	}
	return nil
}

// ensurePublicPolicy 确保 public/ 前缀的对象可公开访问
func (o *OSSClient) ensurePublicPolicy() error {
	policy := fmt.Sprintf(`{
		"Version": "1",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": ["*"],
				"Action": ["oss:GetObject"],
				"Resource": ["acs:oss:*:*:%s/%s*"]
			}
		]
	}`, o.bucketName, PublicPrefix)

	return o.client.SetBucketPolicy(o.bucketName, policy)
}

// PutObject 上传对象到OSS
func (o *OSSClient) PutObject(_ context.Context, key string, reader io.Reader, size int64, mimeType string, _ ...PutOption) error {
	// OSS 权限由 Bucket Policy 控制，PutOption 在此无需处理
	if err := o.bucket.PutObject(key, reader, oss.ContentType(mimeType), oss.ContentLength(size)); err != nil {
		return errs.WrapInternal(err, "上传文件失败")
	}
	return nil
}

// GetURL 获取资源访问URL
func (o *OSSClient) GetURL(_ context.Context, key string, opts ...URLOption) (string, error) {
	options := ApplyURLOptions(opts...)

	// 检查对象是否存在
	exists, err := o.bucket.IsObjectExist(key)
	if err != nil {
		return "", errs.WrapInternal(err, "检查文件是否存在失败")
	}
	if !exists {
		return "", errs.NewBusiness("文件不存在").WithReason(errs.ReasonNotFound)
	}

	// 公开对象：返回直接URL（永不过期）
	if IsPublicKey(key) {
		return o.buildPublicURL(key, options), nil
	}

	// 私有对象：返回预签名URL
	return o.buildSignedURL(key, options)
}

// buildPublicURL 构建公开访问URL
func (o *OSSClient) buildPublicURL(key string, opts *URLOptions) string {
	var baseURL string
	if o.publicURL != "" {
		baseURL = fmt.Sprintf("%s/%s", o.publicURL, key)
	} else {
		baseURL = fmt.Sprintf("https://%s.%s/%s", o.bucketName, o.client.Config.Endpoint, key)
	}

	// 添加图片处理参数（水印等）
	if opts.Watermark != "" {
		return fmt.Sprintf("%s?x-oss-process=%s", baseURL, url.QueryEscape(opts.Watermark))
	}

	return baseURL
}

// buildSignedURL 构建预签名URL
func (o *OSSClient) buildSignedURL(key string, opts *URLOptions) (string, error) {
	var ossOptions []oss.Option

	// 下载模式：设置 Content-Disposition
	if opts.Download && opts.Filename != "" {
		disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(opts.Filename))
		ossOptions = append(ossOptions, oss.ResponseContentDisposition(disposition))
	}

	// 添加图片处理参数（水印等）
	if opts.Watermark != "" {
		ossOptions = append(ossOptions, oss.Process(opts.Watermark))
	}

	signedURL, err := o.bucket.SignURL(key, oss.HTTPGet, int64(o.urlExpires.Seconds()), ossOptions...)
	if err != nil {
		return "", errs.WrapInternal(err, "生成预签名URL失败")
	}

	// 应用CDN地址
	return o.applyPublicURL(signedURL), nil
}

// applyPublicURL 如果配置了公开URL（CDN），替换域名
func (o *OSSClient) applyPublicURL(signedURL string) string {
	if o.publicURL == "" {
		return signedURL
	}

	parsedURL, err := url.Parse(signedURL)
	if err != nil {
		return signedURL
	}

	parsedPublicURL, err := url.Parse(o.publicURL)
	if err != nil {
		return signedURL
	}

	parsedURL.Scheme = parsedPublicURL.Scheme
	parsedURL.Host = parsedPublicURL.Host
	return parsedURL.String()
}

// Exists 检查对象是否存在
func (o *OSSClient) Exists(_ context.Context, key string) (bool, error) {
	exists, err := o.bucket.IsObjectExist(key)
	if err != nil {
		return false, errs.WrapInternal(err, "检查文件是否存在失败")
	}
	return exists, nil
}

// DeleteObject 删除对象
func (o *OSSClient) DeleteObject(_ context.Context, key string) error {
	err := o.bucket.DeleteObject(key)
	if err != nil {
		return errs.WrapInternal(err, "删除文件失败")
	}
	return nil
}

// StatObject 获取对象信息
func (o *OSSClient) StatObject(_ context.Context, key string) (*ObjectInfo, error) {
	header, err := o.bucket.GetObjectMeta(key)
	if err != nil {
		var ossErr oss.ServiceError
		if errors.As(err, &ossErr) && ossErr.Code == "NoSuchKey" {
			return nil, errs.NewBusiness("文件不存在").WithReason(errs.ReasonNotFound)
		}
		return nil, errs.WrapInternal(err, "获取文件信息失败")
	}

	// 解析文件大小
	size, _ := strconv.ParseInt(header.Get("Content-Length"), 10, 64)

	// 解析最后修改时间
	var lastModified time.Time
	if lm := header.Get("Last-Modified"); lm != "" {
		lastModified, _ = time.Parse(time.RFC1123, lm)
	}

	return &ObjectInfo{
		Key:          key,
		Size:         size,
		MimeType:     header.Get("Content-Type"),
		LastModified: lastModified,
		ETag:         header.Get("ETag"),
	}, nil
}
