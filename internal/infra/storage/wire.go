// Package storage provides storage abstraction for different storage backends.
package storage

import (
	"fmt"

	"github.com/google/wire"

	"github.com/xelarion/go-layout/internal/infra/config"
)

// ProviderSet provides storage dependencies.
var ProviderSet = wire.NewSet(
	ProvideStorage,
)

// ProvideStorage 提供Storage实例，根据配置选择存储后端
func ProvideStorage(cfg *config.Storage) (Storage, error) {
	switch cfg.Provider {
	case "oss":
		return NewOSSClient(cfg)
	case "cos":
		return NewCOSClient(cfg)
	case "minio", "":
		return NewMinIOClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}
