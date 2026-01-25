package service

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// ImportService handles import related services.
type ImportService struct {
	importUseCase *usecase.ImportUseCase
	logger        *zap.Logger
}

// NewImportService creates a new ImportService.
func NewImportService(importUseCase *usecase.ImportUseCase, logger *zap.Logger) *ImportService {
	return &ImportService{
		importUseCase: importUseCase,
		logger:        logger,
	}
}

func (s *ImportService) ImportCustomers(ctx context.Context, req *types.ImportCustomersReq) (*types.ImportCustomersResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	params := usecase.ImportCommonParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    0,
		},
		BlobKey: req.Files.Keys()[0],
	}
	err := s.importUseCase.ImportCustomers(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.ImportCustomersResp{}, nil
}

func (s *ImportService) ImportEquipments(ctx context.Context, req *types.ImportEquipmentsReq) (*types.ImportEquipmentsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	params := usecase.ImportCommonParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    0,
		},
		BlobKey: req.Files.Keys()[0],
	}
	if err := s.importUseCase.ImportEquipments(ctx, params); err != nil {
		return nil, err
	}

	return &types.ImportEquipmentsResp{}, nil
}

func (s *ImportService) ImportEquipmentTypes(ctx context.Context, req *types.ImportEquipmentTypesReq) (*types.ImportEquipmentTypesResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	params := usecase.ImportCommonParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    0,
		},
		BlobKey: req.Files.Keys()[0],
	}
	err := s.importUseCase.ImportEquipmentTypes(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.ImportEquipmentTypesResp{}, nil
}

func (s *ImportService) ImportSuppliers(ctx context.Context, req *types.ImportSuppliersReq) (*types.ImportSuppliersResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	params := usecase.ImportCommonParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    0,
		},
		BlobKey: req.Files.Keys()[0],
	}
	err := s.importUseCase.ImportSuppliers(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.ImportSuppliersResp{}, nil
}

func (s *ImportService) DownloadImportTemplate(ctx context.Context, req *types.DownloadImportTemplateReq) ([]byte, string, error) {
	tmpDir := os.TempDir()
	var tmpFileName string
	var tmpFilePath string

	switch req.Type {
	case "customer":
		tmpFileName = "客户导入模板.xlsx"
		tmpFilePath = filepath.Join(tmpDir, tmpFileName)
		err := s.importUseCase.DownloadImportCustomersTemplate(ctx, tmpFilePath)
		if err != nil {
			return nil, "", err
		}
	case "supplier":
		tmpFileName = "供应商导入模板.xlsx"
		tmpFilePath = filepath.Join(tmpDir, tmpFileName)
		err := s.importUseCase.DownloadImportSuppliersTemplate(ctx, tmpFilePath)
		if err != nil {
			return nil, "", err
		}
	case "equipment_type":
		tmpFileName = "设备类型导入模板.xlsx"
		tmpFilePath = filepath.Join(tmpDir, tmpFileName)
		err := s.importUseCase.DownloadImportEquipmentTypesTemplate(ctx, tmpFilePath)
		if err != nil {
			return nil, "", err
		}
	case "equipment":
		tmpFileName = "设备导入模板.xlsx"
		tmpFilePath = filepath.Join(tmpDir, tmpFileName)
		err := s.importUseCase.DownloadImportEquipmentsTemplate(ctx, tmpFilePath)
		if err != nil {
			return nil, "", err
		}
	default:
		return nil, "", errs.NewBusiness("不支持的模板类型")
	}

	data, err := os.ReadFile(tmpFilePath)
	if err != nil {
		return nil, "", err
	}
	_ = os.Remove(tmpFilePath)

	return data, tmpFileName, nil
}
