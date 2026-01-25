package enum

import "github.com/xelarion/go-layout/pkg/errs"

// BatchDetailStatus 表示批次详情状态
type BatchDetailStatus int8

const (
	BatchDetailStatusPending    BatchDetailStatus = iota // 待处理
	BatchDetailStatusProcessing                          // 处理中
	BatchDetailStatusSuccess                             // 成功
	BatchDetailStatusFailed                              // 失败
)

var BatchDetailStatusText = map[BatchDetailStatus]string{
	BatchDetailStatusPending:    "待处理",
	BatchDetailStatusProcessing: "处理中",
	BatchDetailStatusSuccess:    "成功",
	BatchDetailStatusFailed:     "失败",
}

var BatchDetailStatusSlug = map[BatchDetailStatus]string{
	BatchDetailStatusPending:    "pending",
	BatchDetailStatusProcessing: "processing",
	BatchDetailStatusSuccess:    "success",
	BatchDetailStatusFailed:     "failed",
}

func GetBatchDetailStatus(slug string) (BatchDetailStatus, error) {
	for k, v := range BatchDetailStatusSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid batch detail status slug")
}
