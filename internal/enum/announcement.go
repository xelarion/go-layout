package enum

import "github.com/xelarion/go-layout/pkg/errs"

// AnnouncementStatus 公告状态
type AnnouncementStatus int8

const (
	AnnouncementStatusDraft     AnnouncementStatus = iota // 暂存
	AnnouncementStatusPublished                           // 已发布
)

var AnnouncementStatusText = map[AnnouncementStatus]string{
	AnnouncementStatusDraft:     "暂存",
	AnnouncementStatusPublished: "已发布",
}

var AnnouncementStatusSlug = map[AnnouncementStatus]string{
	AnnouncementStatusDraft:     "draft",
	AnnouncementStatusPublished: "published",
}

func GetAnnouncementStatus(slug string) (AnnouncementStatus, error) {
	for k, v := range AnnouncementStatusSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("无效的公告状态")
}
