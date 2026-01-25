package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/pkg/errs"
)

// COSClient 腾讯云COS存储客户端，实现Storage接口
type COSClient struct {
	client     *cos.Client
	bucketName string
	region     string
	appID      string
	secretID   string
	secretKey  string
	urlExpires time.Duration
	publicURL  string
}

// NewCOSClient 创建腾讯云COS客户端
func NewCOSClient(cfg *config.Storage) (*COSClient, error) {
	if cfg.COSRegion == "" {
		return nil, fmt.Errorf("cos region is required")
	}

	if cfg.COSAppID == "" {
		return nil, fmt.Errorf("cos app id is required")
	}

	// 验证 bucket 名称格式：必须以 -<appid> 结尾
	expectedSuffix := "-" + cfg.COSAppID
	if !strings.HasSuffix(cfg.Bucket, expectedSuffix) {
		return nil, fmt.Errorf("cos bucket name must end with '-%s', got %q (expected format: <name>-%s)", cfg.COSAppID, cfg.Bucket, cfg.COSAppID)
	}

	// 构建 Bucket URL: https://<bucket>.cos.<region>.myqcloud.com
	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.COSRegion))

	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COSSecretID,
			SecretKey: cfg.COSSecretKey,
		},
	})

	c := &COSClient{
		client:     client,
		bucketName: cfg.Bucket,
		region:     cfg.COSRegion,
		appID:      cfg.COSAppID,
		secretID:   cfg.COSSecretID,
		secretKey:  cfg.COSSecretKey,
		urlExpires: cfg.URLExpires,
		publicURL:  cfg.PublicURL,
	}

	ctx := context.Background()

	// 确保 bucket 存在
	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}

	// 设置 public/ 前缀的公开读策略
	if err := c.ensurePublicPolicy(ctx); err != nil {
		return nil, fmt.Errorf("failed to set public policy: %w", err)
	}

	return c, nil
}

// ensureBucket 确保 bucket 存在
func (c *COSClient) ensureBucket(ctx context.Context) error {
	ok, err := c.client.Bucket.IsExist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if ok {
		return nil
	}

	// 尝试创建 bucket，如果没有权限则返回错误提示用户手动创建
	if _, err := c.client.Bucket.Put(ctx, nil); err != nil {
		return fmt.Errorf("bucket %q create failed: %w, or you can create it manually in Tencent Cloud console", c.bucketName, err)
	}
	return nil
}

// ensurePublicPolicy 确保 public/ 前缀的对象可公开访问
func (c *COSClient) ensurePublicPolicy(ctx context.Context) error {
	// COS Resource 格式: qcs::cos:<region>:uid/<appid>:<bucket>/<key>
	// 例如: qcs::cos:ap-guangzhou:uid/1250000000:attachments-1250000000/public/*
	statement := cos.BucketStatement{
		Effect:    "Allow",
		Principal: map[string][]string{"qcs": {"*"}},
		Action:    []string{"name/cos:GetObject"},
		Resource:  []string{fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s*", c.region, c.appID, c.bucketName, PublicPrefix)},
	}

	_, err := c.client.Bucket.PutPolicy(ctx, &cos.BucketPutPolicyOptions{
		Version:   "2.0",
		Statement: []cos.BucketStatement{statement},
	})
	return err
}

// PutObject 上传对象到COS
func (c *COSClient) PutObject(ctx context.Context, key string, reader io.Reader, size int64, mimeType string, _ ...PutOption) error {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType:   mimeType,
			ContentLength: size,
		},
	}

	if _, err := c.client.Object.Put(ctx, key, reader, opt); err != nil {
		return errs.WrapInternal(err, "上传文件失败")
	}
	return nil
}

// GetURL 获取资源访问URL
func (c *COSClient) GetURL(ctx context.Context, key string, opts ...URLOption) (string, error) {
	options := ApplyURLOptions(opts...)

	// 检查对象是否存在
	exists, err := c.Exists(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errs.NewBusiness("文件不存在").WithReason(errs.ReasonNotFound)
	}

	// 公开对象：返回直接URL
	if IsPublicKey(key) {
		return c.buildPublicURL(key, options), nil
	}

	// 私有对象：返回预签名URL
	return c.buildPresignedURL(ctx, key, options)
}

// buildPublicURL 构建公开访问URL
func (c *COSClient) buildPublicURL(key string, opts *URLOptions) string {
	var baseURL string
	if c.publicURL != "" {
		baseURL = fmt.Sprintf("%s/%s", c.publicURL, key)
	} else {
		baseURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", c.bucketName, c.region, key)
	}

	// COS 图片处理参数（水印等），格式如: imageMogr2/watermark/...
	if opts.Watermark != "" {
		return fmt.Sprintf("%s?%s", baseURL, url.QueryEscape(opts.Watermark))
	}

	return baseURL
}

// buildPresignedURL 构建预签名URL
func (c *COSClient) buildPresignedURL(ctx context.Context, key string, opts *URLOptions) (string, error) {
	presignedURL, err := c.client.Object.GetPresignedURL(ctx, http.MethodGet, key, c.secretID, c.secretKey, c.urlExpires, nil)
	if err != nil {
		return "", errs.WrapInternal(err, "生成预签名URL失败")
	}

	// 下载模式：添加 response-content-disposition 参数
	if opts.Download && opts.Filename != "" {
		disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(opts.Filename))
		q := presignedURL.Query()
		q.Set("response-content-disposition", disposition)
		presignedURL.RawQuery = q.Encode()
	}

	// COS 图片处理参数（水印等），直接追加到 URL
	if opts.Watermark != "" {
		if presignedURL.RawQuery != "" {
			presignedURL.RawQuery += "&" + opts.Watermark
		} else {
			presignedURL.RawQuery = opts.Watermark
		}
	}

	return c.applyPublicURL(presignedURL), nil
}

// applyPublicURL 如果配置了公开URL（CDN），替换域名
func (c *COSClient) applyPublicURL(presignedURL *url.URL) string {
	if c.publicURL == "" {
		return presignedURL.String()
	}

	parsedPublicURL, err := url.Parse(c.publicURL)
	if err != nil {
		return presignedURL.String()
	}

	presignedURL.Scheme = parsedPublicURL.Scheme
	presignedURL.Host = parsedPublicURL.Host
	return presignedURL.String()
}

// Exists 检查对象是否存在
func (c *COSClient) Exists(ctx context.Context, key string) (bool, error) {
	ok, err := c.client.Object.IsExist(ctx, key)
	if err != nil {
		return false, errs.WrapInternal(err, "检查文件是否存在失败")
	}
	return ok, nil
}

// DeleteObject 删除对象
func (c *COSClient) DeleteObject(ctx context.Context, key string) error {
	if _, err := c.client.Object.Delete(ctx, key); err != nil {
		return errs.WrapInternal(err, "删除文件失败")
	}
	return nil
}

// StatObject 获取对象信息
func (c *COSClient) StatObject(ctx context.Context, key string) (*ObjectInfo, error) {
	resp, err := c.client.Object.Head(ctx, key, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, errs.NewBusiness("文件不存在").WithReason(errs.ReasonNotFound)
		}
		return nil, errs.WrapInternal(err, "获取文件信息失败")
	}

	// 解析最后修改时间
	var lastModified time.Time
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		lastModified, _ = time.Parse(time.RFC1123, lm)
	}

	return &ObjectInfo{
		Key:          key,
		Size:         resp.ContentLength,
		MimeType:     resp.Header.Get("Content-Type"),
		LastModified: lastModified,
		ETag:         resp.Header.Get("ETag"),
	}, nil
}
