package model

import (
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

type StorageAttachment struct {
	gen.StorageAttachment

	Name enum.StorageAttachmentNameType `gorm:"column:name;type:character varying(255);not null;comment:名称, 如: file, avatar (区分同一条记录不同类型的附件)" json:"name"` // 名称, 如: file, avatar (区分同一条记录不同类型的附件)

	StorageBlob StorageBlob `gorm:"foreignKey:StorageBlobID;references:ID;<-:create" json:"storage_blob"`
}

// GetResourceType 实现 Resource 接口
func (s *StorageAttachment) GetResourceType() string {
	return "storage_attachment"
}

// GetResourceID 实现 Resource 接口
func (s *StorageAttachment) GetResourceID() uint {
	return s.ID
}

func (s *StorageAttachment) AfterSave(tx *gorm.DB) (err error) {
	// 关联创建时唯一约束冲突应 return err
	// 这里不使用 Clauses(clause.OnConflict{DoNothing: false}) 而使用 AfterSave 检查
	if s.StorageBlobID == 0 {
		return errs.NewInternal("storage_blob_id must grater than zero")
	}
	return
}
