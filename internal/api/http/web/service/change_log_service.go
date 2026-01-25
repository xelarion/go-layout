package service

import (
	"context"
	"encoding/json"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
)

// ChangeLogService handles change_log-related services.
type ChangeLogService struct {
	*BaseService
	changeLogUseCase *usecase.ChangeLogUseCase
	exportUseCase    *usecase.ExportUseCase
}

// NewChangeLogService creates a new ChangeLogService.
func NewChangeLogService(
	baseService *BaseService,
	changeLogUseCase *usecase.ChangeLogUseCase,
	exportUseCase *usecase.ExportUseCase,
) *ChangeLogService {
	return &ChangeLogService{
		BaseService:      baseService,
		changeLogUseCase: changeLogUseCase,
		exportUseCase:    exportUseCase,
	}
}

// ListChangeLogs lists change_logs.
func (s *ChangeLogService) ListChangeLogs(ctx context.Context, req *types.ListChangeLogsReq) (*types.ListChangeLogsResp, error) {
	filters := map[string]any{
		"resource_type": req.ResourceType,
		"resource_id":   req.ResourceID,
	}

	changeLogs, _, err := s.changeLogUseCase.List(ctx, filters, -1, -1, req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]*types.ListChangeLogsRespResult, 0, len(changeLogs))
	for _, changeLog := range changeLogs {
		details := make([]*types.ListChangeLogsRespResultDetail, 0, len(changeLog.Details))
		for _, detail := range changeLog.Details {
			respDetail := &types.ListChangeLogsRespResultDetail{
				ID:         detail.ID,
				FieldName:  detail.FieldName,
				FieldLabel: detail.FieldLabel,
				OldValue:   detail.OldValue,
				NewValue:   detail.NewValue,
			}

			// 根据字段类型和关联类型判断是否需要解析为数组
			if detail.FieldType == enum.ChangeLogDetailFieldTypeRelation && detail.RelationType == enum.ChangeLogDetailRelationTypeMany {
				// 解析 RelationMeta
				meta := detail.RelationMeta.Data()
				if len(meta.Schema) > 0 {
					schema := make([]types.ChangeLogDetailSchemaFieldResp, 0, len(meta.Schema))
					for _, field := range meta.Schema {
						schema = append(schema, types.ChangeLogDetailSchemaFieldResp{
							Name:  field.Name,
							Label: field.Label,
						})
					}
					respDetail.RelationSchema = schema
				}
				respDetail.RelationGroupBy = meta.GroupBy

				// 解析 JSON 字符串为数组
				if detail.OldValue != "" && detail.OldValue != "[]" {
					var oldRelations []map[string]string
					if err := json.Unmarshal([]byte(detail.OldValue), &oldRelations); err == nil {
						respDetail.OldRelations = oldRelations
						respDetail.OldValue = ""
					}
				}
				if detail.NewValue != "" && detail.NewValue != "[]" {
					var newRelations []map[string]string
					if err := json.Unmarshal([]byte(detail.NewValue), &newRelations); err == nil {
						respDetail.NewRelations = newRelations
						respDetail.NewValue = ""
					}
				}
			}

			details = append(details, respDetail)
		}

		result := &types.ListChangeLogsRespResult{
			ID:              changeLog.ID,
			CreatedAt:       util.FormatDateTime(changeLog.CreatedAt),
			OperatorDisplay: changeLog.OperatorDisplay,
			Action:          changeLog.Action.Text(),
			ActionSlug:      changeLog.Action.Slug(),
			Details:         details,
		}
		respResults = append(respResults, result)
	}

	return &types.ListChangeLogsResp{
		Results: respResults,
	}, nil
}
