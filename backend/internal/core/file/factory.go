package file

import (
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"go.uber.org/zap"
)

func NewFileService(cfg config.FileConfig, logger *zap.Logger) OSS {
	switch cfg.Driver {
	case "local":
		return NewLocalDriver(cfg.Local, logger)
	case "minio":
		// ✨ 注册 MinIO
		return NewMinioDriver(cfg.Minio, logger)
	case "cos":
		driver, err := NewCOSDriver(cfg.COS)
		if err != nil {
			logger.Panic("cos init failed", zap.Error(err))
		}
		return driver
	default:
		return NewLocalDriver(cfg.Local, logger)
	}
}
