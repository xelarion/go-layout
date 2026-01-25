package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// StorageService 文件存储服务
type StorageService struct {
	storageUC *usecase.StorageUseCase
}

// NewStorageService 创建StorageService实例
func NewStorageService(storageUC *usecase.StorageUseCase) *StorageService {
	return &StorageService{
		storageUC: storageUC,
	}
}

// Upload 处理文件上传
func (s *StorageService) Upload(ctx context.Context, req *types.UploadFileReq) (*types.UploadFileResp, error) {
	// 打开文件流
	file, err := req.File.Open()
	if err != nil {
		return nil, errs.WrapBusiness(err, "打开文件失败")
	}
	defer file.Close()

	// 构造上传参数
	params := usecase.UploadParams{
		File:     file,
		Filename: req.File.Filename,
		Size:     req.File.Size,
		MimeType: req.File.Header.Get("Content-Type"),
		Public:   req.Public,
	}

	// 调用usecase上传
	result, err := s.storageUC.Upload(ctx, params)
	if err != nil {
		return nil, err
	}

	// 构造响应
	return &types.UploadFileResp{
		Key:      result.Key,
		Filename: result.Filename,
		Size:     result.Size,
		MimeType: result.MimeType,
		Url:      result.Url,
	}, nil
}

// GetDownloadURL 获取文件下载URL（带权限检查）
func (s *StorageService) GetDownloadURL(ctx context.Context, req *types.GetDownloadURLReq) (*types.GetDownloadURLResp, error) {
	url, err := s.storageUC.GenerateDownloadURL(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.GetDownloadURLResp{
		URL: url,
	}, nil
}
