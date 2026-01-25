package model

import "github.com/xelarion/go-layout/internal/model/gen"

type StorageBlob struct {
	gen.StorageBlob
}

func (s *StorageBlob) GetResourceType() string {
	return "storage_blob"
}

func (s *StorageBlob) GetResourceID() uint {
	return s.ID
}
