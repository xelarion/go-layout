package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/permission"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// BaseService 提供 Service 层通用的数据权限处理方法
type BaseService struct {
	roleUseCase *usecase.RoleUseCase
}

// NewBaseService 创建基础 Service
func NewBaseService(roleUseCase *usecase.RoleUseCase) *BaseService {
	return &BaseService{roleUseCase: roleUseCase}
}

// ApplyDataScope 应用数据权限过滤到列表查询
// 自动从 context 获取 ReqPermissions，根据角色的 DataScopes 配置添加过滤条件
// record 参数用于获取字段映射，传入空实例即可（如 &model.User{}）
// 如果 ReqPermissions 为空（内部调用，如中间件），则跳过检查
func (b *BaseService) ApplyDataScope(ctx context.Context, filters map[string]any, record permission.DataScopeRecord) error {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 没有 ReqPermissions，跳过数据权限检查
	if len(current.ReqPermissions) == 0 {
		return nil
	}

	// 查询角色的数据权限配置
	role, err := b.roleUseCase.GetByID(ctx, current.RoleID)
	if err != nil {
		return err
	}

	// 应用数据权限过滤（传入 record 用于字段映射）
	filter := permission.NewDataScopeFilter(ctx, role.Role, record)
	if !filter.Apply(ctx, filters) {
		return errs.NewBusiness("data access denied").
			WithReason(errs.ReasonForbidden)
	}

	return nil
}

// CheckRecordAccess 检查是否有权访问单条记录
// 自动从 context 获取 ReqPermissions，根据角色的 DataScopes 配置检查记录权限
// 如果 ReqPermissions 为空（内部调用，如中间件），则跳过检查
func (b *BaseService) CheckRecordAccess(ctx context.Context, record permission.DataScopeRecord) error {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 内部调用（如 AuthMiddleware）没有 ReqPermissions，跳过数据权限检查
	if len(current.ReqPermissions) == 0 {
		return nil
	}

	// 查询角色的数据权限配置
	role, err := b.roleUseCase.GetByID(ctx, current.RoleID)
	if err != nil {
		return err
	}

	// 检查记录访问权限（record 同时用于检查和字段映射）
	filter := permission.NewDataScopeFilter(ctx, role.Role, record)
	if !filter.CheckRecordAccess(ctx, record) {
		return errs.NewBusiness("data access denied").
			WithReason(errs.ReasonForbidden)
	}

	return nil
}

// NewExportResp 从 ExportResult 创建 ExportResp
func (b *BaseService) NewExportResp(result *usecase.ExportResult) *types.ExportResp {
	return &types.ExportResp{
		ID:          result.TaskID,
		IsAsync:     result.IsAsync,
		TotalCount:  result.TotalCount,
		DownloadURL: result.DownloadURL,
	}
}

// NewCommonUploadedFileResp 从 UploadResult 创建 CommonUploadedFileResp
func (b *BaseService) NewCommonUploadedFileResp(results []usecase.UploadResult) []types.CommonUploadedFileResp {
	var resp []types.CommonUploadedFileResp
	for _, result := range results {
		resp = append(resp, types.CommonUploadedFileResp{
			Key:      result.Key,
			Filename: result.Filename,
			Size:     result.Size,
			MimeType: result.MimeType,
			Url:      result.Url,
		})
	}
	return resp
}
