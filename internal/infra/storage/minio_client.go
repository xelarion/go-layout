package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/pkg/errs"
)

// MinIOClient MinIO存储客户端，实现Storage接口
type MinIOClient struct {
	client     *minio.Client
	bucket     string
	urlExpires time.Duration
	publicURL  string // 公开访问地址（CDN）
}

// NewMinIOClient 创建MinIO客户端
func NewMinIOClient(cfg *config.Storage) (*MinIOClient, error) {
	if cfg.MinIOEndpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}

	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKeyID, cfg.MinIOSecretAccessKey, ""),
		Secure: cfg.MinIOUseSSL,
		Region: cfg.MinIORegion,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	m := &MinIOClient{
		client:     client,
		bucket:     cfg.Bucket,
		urlExpires: cfg.URLExpires,
		publicURL:  cfg.PublicURL,
	}

	ctx := context.Background()

	// 确保 bucket 存在
	if err := m.ensureBucket(ctx, cfg.MinIORegion); err != nil {
		return nil, err
	}

	// 设置 public/ 前缀的公开读策略
	if err := m.ensurePublicPolicy(ctx); err != nil {
		return nil, fmt.Errorf("failed to set public policy: %w", err)
	}

	return m, nil
}

// ensureBucket 确保 bucket 存在
func (m *MinIOClient) ensureBucket(ctx context.Context, region string) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if exists {
		return nil
	}

	if err := m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		return fmt.Errorf("bucket %q create failed: %w, or you can create it manually in MinIO console", m.bucket, err)
	}
	return nil
}

// ensurePublicPolicy 确保 public/ 前缀的对象可公开访问
func (m *MinIOClient) ensurePublicPolicy(ctx context.Context) error {
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":    "Allow",
				"Principal": map[string]any{"AWS": []string{"*"}},
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/%s*", m.bucket, PublicPrefix)},
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	return m.client.SetBucketPolicy(ctx, m.bucket, string(policyJSON))
}

// PutObject 上传对象到MinIO
func (m *MinIOClient) PutObject(ctx context.Context, key string, reader io.Reader, size int64, mimeType string, _ ...PutOption) error {
	// MinIO 权限由 Bucket Policy 控制，PutOption 在此无需处理
	_, err := m.client.PutObject(ctx, m.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return errs.WrapInternal(err, "上传文件失败")
	}
	return nil
}

// GetURL 获取资源访问URL
func (m *MinIOClient) GetURL(ctx context.Context, key string, opts ...URLOption) (string, error) {
	options := ApplyURLOptions(opts...)

	// 检查对象是否存在
	exists, err := m.Exists(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errs.NewBusiness("文件不存在").WithReason(errs.ReasonNotFound)
	}

	// 公开对象：返回直接URL（永不过期）
	if IsPublicKey(key) {
		return m.buildPublicURL(key), nil
	}

	// 私有对象：返回预签名URL
	return m.buildPresignedURL(ctx, key, options)
}

// buildPublicURL 构建公开访问URL
func (m *MinIOClient) buildPublicURL(key string) string {
	if m.publicURL != "" {
		return fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucket, key)
	}
	return fmt.Sprintf("%s/%s/%s", m.client.EndpointURL().String(), m.bucket, key)
}

// buildPresignedURL 构建预签名URL
func (m *MinIOClient) buildPresignedURL(ctx context.Context, key string, opts *URLOptions) (string, error) {
	reqParams := url.Values{}

	// 下载模式：设置 Content-Disposition
	if opts.Download && opts.Filename != "" {
		disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(opts.Filename))
		reqParams.Set("response-content-disposition", disposition)
	}

	// MinIO 不支持水印，忽略 opts.Watermark

	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, key, m.urlExpires, reqParams)
	if err != nil {
		return "", errs.WrapInternal(err, "生成预签名URL失败")
	}

	return m.applyPublicURL(presignedURL), nil
}

// applyPublicURL 如果配置了公开URL（CDN），替换域名
func (m *MinIOClient) applyPublicURL(presignedURL *url.URL) string {
	if m.publicURL == "" {
		return presignedURL.String()
	}

	parsedPublicURL, err := url.Parse(m.publicURL)
	if err != nil {
		return presignedURL.String()
	}

	presignedURL.Scheme = parsedPublicURL.Scheme
	presignedURL.Host = parsedPublicURL.Host
	return presignedURL.String()
}

// Exists 检查对象是否存在
func (m *MinIOClient) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, errs.WrapInternal(err, "检查文件是否存在失败")
	}
	return true, nil
}

// DeleteObject 删除对象
func (m *MinIOClient) DeleteObject(ctx context.Context, key string) error {
	err := m.client.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return errs.WrapInternal(err, "删除文件失败")
	}
	return nil
}

// StatObject 获取对象信息
func (m *MinIOClient) StatObject(ctx context.Context, key string) (*ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, m.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, errs.NewBusiness("文件不存在").WithReason(errs.ReasonNotFound)
		}
		return nil, errs.WrapInternal(err, "获取文件信息失败")
	}

	return &ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		MimeType:     info.ContentType,
		LastModified: info.LastModified,
		ETag:         info.ETag,
	}, nil
}
