// Package usecase contains business logic.
package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/infra/storage"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

type StorageAttachment struct {
	*model.StorageAttachment
}

// StorageUseCase 文件存储业务逻辑
type StorageUseCase struct {
	tx                    Transaction
	storage               storage.Storage
	storageBlobRepo       StorageBlobRepository
	storageAttachmentRepo StorageAttachmentRepository
	config                *config.Storage
	logger                *zap.Logger
}

// NewStorageUseCase 创建StorageUseCase实例
func NewStorageUseCase(
	tx Transaction,
	storage storage.Storage,
	storageBlobRepo StorageBlobRepository,
	storageAttachmentRepo StorageAttachmentRepository,
	config *config.Storage,
	logger *zap.Logger,
) *StorageUseCase {
	return &StorageUseCase{
		tx:                    tx,
		storage:               storage,
		storageBlobRepo:       storageBlobRepo,
		storageAttachmentRepo: storageAttachmentRepo,
		config:                config,
		logger:                logger.Named("storage-usecase"),
	}
}

// UploadParams 上传参数
type UploadParams struct {
	File     io.Reader // 文件流
	Filename string    // 原始文件名
	Size     int64     // 文件大小
	MimeType string    // MIME类型
	Public   bool      // 是否公开访问（适用于通知公告等场景）
}

// UploadResult 上传结果
type UploadResult struct {
	Key      string // 文件唯一标识
	Filename string // 原始文件名
	Size     int64  // 文件大小
	MimeType string // MIME类型
	Url      string // 资源链接
}

// Upload 上传文件并保存元数据
func (uc *StorageUseCase) Upload(ctx context.Context, params UploadParams) (*UploadResult, error) {
	// 1. 验证文件
	if err := uc.validateFile(params); err != nil {
		return nil, err
	}

	// 2. 生成唯一key
	key := uc.generateKey(params.Filename, params.Public)

	// 3. 上传到存储
	var opts []storage.PutOption
	if params.Public {
		opts = append(opts, storage.WithPutPublic())
	}
	if err := uc.storage.PutObject(ctx, key, params.File, params.Size, params.MimeType, opts...); err != nil {
		return nil, err
	}

	// 4. 保存元数据到数据库（事务）
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		blob := &model.StorageBlob{
			StorageBlob: gen.StorageBlob{
				Key:      key,
				Filename: params.Filename,
				MimeType: params.MimeType,
				ByteSize: params.Size,
			},
		}
		if err := uc.storageBlobRepo.Create(ctx, blob); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		// 数据库保存失败，回滚：删除已上传的文件
		_ = uc.storage.DeleteObject(context.Background(), key)
		return nil, err
	}

	// 生成资源链接
	url, err := uc.storage.GetURL(ctx, key)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		Key:      key,
		Filename: params.Filename,
		Size:     params.Size,
		MimeType: params.MimeType,
		Url:      url,
	}, nil
}

// GenerateDownloadURL 生成文件下载URL
func (uc *StorageUseCase) GenerateDownloadURL(ctx context.Context, key string) (string, error) {
	blob, err := uc.storageBlobRepo.FindByKey(ctx, key)
	if err != nil {
		return "", err
	}

	return uc.storage.GetURL(ctx, blob.Key, storage.WithURLDownload(blob.Filename))
}

// GenerateResourceURL 生成资源展示URL（系统内部调用，用于返回数据时自动包含URL）
// 此方法不检查权限，因为调用方已经通过权限检查获取了资源
func (uc *StorageUseCase) GenerateResourceURL(ctx context.Context, key string) (string, error) {
	blob, err := uc.storageBlobRepo.FindByKey(ctx, key)
	if err != nil {
		return "", err
	}

	return uc.storage.GetURL(ctx, blob.Key)
}

// AttachOne Attaches an attachable to the record.
// newBlobKey 存在则 attach 已经存在的 blob
// newBlobKey 为空则删除当前 attachment
func (uc *StorageUseCase) AttachOne(
	ctx context.Context,
	resource model.Resource,
	name enum.StorageAttachmentNameType,
	newBlobKey string,
) error {
	isOldAttachmentExists := false
	recordType := resource.GetResourceType()
	recordID := resource.GetResourceID()

	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		oldAttachment, err := uc.storageAttachmentRepo.FindByResourceAndName(ctx, recordType, recordID, name)
		if err != nil {
			if errs.IsReason(err, errs.ReasonNotFound) {
				isOldAttachmentExists = false
			} else {
				return err
			}
		} else {
			isOldAttachmentExists = true
		}

		// 没有新的，即删除当前 attachment
		if newBlobKey == "" {
			if isOldAttachmentExists {
				// 删除旧的
				return uc.storageAttachmentRepo.Delete(ctx, oldAttachment.ID)
			}
			return nil
		}

		// 首次上传 或者 更换了
		if !isOldAttachmentExists || newBlobKey != oldAttachment.StorageBlob.Key {
			if isOldAttachmentExists {
				// 删除旧的
				if err = uc.storageAttachmentRepo.Delete(ctx, oldAttachment.ID); err != nil {
					return err
				}
			}

			newBlob, err := uc.storageBlobRepo.FindByKey(ctx, newBlobKey)
			if err != nil {
				return err
			}

			// 创建新的 attachment
			newStorageAttachment := &model.StorageAttachment{
				StorageAttachment: gen.StorageAttachment{
					ResourceType:  recordType,
					ResourceID:    recordID,
					StorageBlobID: newBlob.ID,
				},
				Name: name,
			}
			return uc.storageAttachmentRepo.Create(ctx, newStorageAttachment)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 没变
	return nil
}

// AttachMany Attaches attachable to the record.
// newBlobKeys 存在则 attach 已经存在的 blobs
// newBlobKeys 为空则删除当前 attachments
func (uc *StorageUseCase) AttachMany(
	ctx context.Context,
	resource model.Resource,
	name enum.StorageAttachmentNameType,
	newBlobKeys []string,
) error {
	recordType := resource.GetResourceType()
	recordID := resource.GetResourceID()

	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		oldAttachments, err := uc.storageAttachmentRepo.ListByResourceAndName(ctx, recordType, recordID, name)
		if err != nil {
			return err
		}

		oldBlobKeys := make([]string, len(oldAttachments))
		for i, attachment := range oldAttachments {
			oldBlobKeys[i] = attachment.StorageBlob.Key
		}

		// deleteKeys 需要删除的 key
		// createKeys 需要创建的 key
		deleteKeys, createKeys := util.DifferenceString(oldBlobKeys, newBlobKeys)
		// 无变化
		if len(deleteKeys) == 0 && len(createKeys) == 0 {
			return nil
		}

		// 删除旧的 attachments
		if len(deleteKeys) > 0 {
			var deleteAttachmentIds []uint
			for _, oldAttachment := range oldAttachments {
				if slices.Contains(deleteKeys, oldAttachment.StorageBlob.Key) {
					deleteAttachmentIds = append(deleteAttachmentIds, oldAttachment.ID)
				}
			}
			err = uc.storageAttachmentRepo.BatchDelete(ctx, deleteAttachmentIds)
			if err != nil {
				return err
			}
		}

		// 创建新的 attachments
		var newAttachments []*model.StorageAttachment
		if len(createKeys) > 0 {
			for _, blobKey := range createKeys {
				blob, err := uc.storageBlobRepo.FindByKey(ctx, blobKey)
				if err != nil {
					return err
				}

				newStorageAttachment := &model.StorageAttachment{
					StorageAttachment: gen.StorageAttachment{
						ResourceType:  recordType,
						ResourceID:    recordID,
						StorageBlobID: blob.ID,
					},
					Name: name,
				}
				newAttachments = append(newAttachments, newStorageAttachment)
			}

			err = uc.storageAttachmentRepo.CreateInBatches(ctx, newAttachments, len(newAttachments))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Detach 分离资源关联的所有 attachments
// blob 和文件由定时清理任务处理
func (uc *StorageUseCase) Detach(ctx context.Context, resource model.Resource) error {
	return uc.storageAttachmentRepo.DeleteByResource(ctx, resource.GetResourceType(), resource.GetResourceID())
}

// FindByResourceAndName 根据资源和名称查找存储附件
func (uc *StorageUseCase) FindByResourceAndName(ctx context.Context, resource model.Resource, name enum.StorageAttachmentNameType) (*StorageAttachment, error) {
	storageAttachment, err := uc.storageAttachmentRepo.FindByResourceAndName(ctx, resource.GetResourceType(), resource.GetResourceID(), name)
	if err != nil {
		return nil, err
	}

	return &StorageAttachment{
		StorageAttachment: storageAttachment,
	}, nil
}

// ListByResourceAndName 根据资源和名称列出存储附件
func (uc *StorageUseCase) ListByResourceAndName(ctx context.Context, resource model.Resource, name enum.StorageAttachmentNameType) ([]*StorageAttachment, error) {
	storageAttachments, err := uc.storageAttachmentRepo.ListByResourceAndName(ctx, resource.GetResourceType(), resource.GetResourceID(), name)
	if err != nil {
		return nil, err
	}

	var attachments []*StorageAttachment
	for _, storageAttachment := range storageAttachments {
		attachments = append(attachments, &StorageAttachment{
			StorageAttachment: storageAttachment,
		})
	}
	return attachments, nil
}

// AttachmentsToUploadResults 转换为上传结果
func (uc *StorageUseCase) AttachmentsToUploadResults(ctx context.Context, storageAttachments []*StorageAttachment) ([]UploadResult, error) {
	var results []UploadResult
	for _, storageAttachment := range storageAttachments {
		url, err := uc.storage.GetURL(ctx, storageAttachment.StorageBlob.Key)
		if err != nil {
			return nil, err
		}

		results = append(results, UploadResult{
			Key:      storageAttachment.StorageBlob.Key,
			Filename: storageAttachment.StorageBlob.Filename,
			Size:     storageAttachment.StorageBlob.ByteSize,
			MimeType: storageAttachment.StorageBlob.MimeType,
			Url:      url,
		})
	}
	return results, nil
}

// validateFile 验证文件
func (uc *StorageUseCase) validateFile(params UploadParams) error {
	// 文件大小限制
	if params.Size > uc.config.MaxFileSize {
		return errs.NewBusiness(fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", uc.config.MaxFileSize))
	}

	if params.Size == 0 {
		return errs.NewBusiness("文件为空")
	}

	// 文件名验证
	if params.Filename == "" {
		return errs.NewBusiness("文件名不能为空")
	}

	// Content-Type验证
	if params.MimeType == "" {
		return errs.NewBusiness("文件类型不能为空")
	}

	// 文件类型白名单验证（支持通配符）
	if !uc.isAllowedMimeType(params.MimeType) {
		return errs.NewBusiness(fmt.Sprintf("不支持的文件类型: %s", params.MimeType))
	}

	return nil
}

// isAllowedMimeType 检查Content-Type是否在允许列表中
// 支持通配符，如 image/*
func (uc *StorageUseCase) isAllowedMimeType(mimeType string) bool {
	for _, allowed := range uc.config.AllowedTypes {
		// * 允许所有
		if allowed == "*" {
			return true
		}

		// 完全匹配
		if allowed == mimeType {
			return true
		}

		// 通配符匹配
		if strings.HasSuffix(allowed, "/*") {
			prefix := strings.TrimSuffix(allowed, "/*")
			if strings.HasPrefix(mimeType, prefix+"/") {
				return true
			}
		}
	}
	return false
}

// generateKey 生成唯一存储路径
// 格式: {uuid}.ext 或 public/{uuid}.ext
func (uc *StorageUseCase) generateKey(filename string, public bool) string {
	key := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(filename))
	if public {
		return storage.PublicPrefix + key
	}
	return key
}

// blobKeyPattern 匹配 blob key 的正则表达式
// 格式: {public_prefix}{uuid}.ext 或 {uuid}.ext
var blobKeyPattern = regexp.MustCompile(`(?:` + regexp.QuoteMeta(storage.PublicPrefix) + `)?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.[a-zA-Z0-9]+`)

// extractBlobKeysFromContent 从富文本内容中提取 blob key
func extractBlobKeysFromContent(content string) []string {
	matches := blobKeyPattern.FindAllString(content, -1)
	seen := make(map[string]struct{})
	var keys []string
	for _, key := range matches {
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
	}
	return keys
}

// AttachContentBlobs 为富文本内容中的 blob 创建 attachment
// 用于公告、通知等富文本场景
func (uc *StorageUseCase) AttachContentBlobs(ctx context.Context, resource model.Resource, content string) error {
	blobKeys := extractBlobKeysFromContent(content)
	return uc.AttachMany(ctx, resource, enum.StorageAttachmentNameTypeContent, blobKeys)
}

// DeleteBlob 删除 blob 及对象存储中的文件
// 仅供定时清理任务使用
func (uc *StorageUseCase) DeleteBlob(ctx context.Context, blobID uint, key string) error {
	if err := uc.storage.DeleteObject(ctx, key); err != nil {
		return err
	}
	return uc.storageBlobRepo.Delete(ctx, blobID)
}
