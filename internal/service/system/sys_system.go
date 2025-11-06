package system

import (
	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"

	"go.uber.org/zap"
)

type IBaseService interface {
	GetSystemConfig() (conf *config.Config, err error)
	GetServerInfo() (server interface{}, err error)
}

type BaseService struct {
	serviceCtx *svc.ServiceContext
}

func NewBaseService(serviceCtx *svc.ServiceContext) IBaseService {
	return &BaseService{serviceCtx: serviceCtx}
}

// GetSystemConfig 读取配置文件
func (s *BaseService) GetSystemConfig() (conf *config.Config, err error) {
	return s.serviceCtx.Config, nil
}

// GetServerInfo 获取服务器信息
func (s *BaseService) GetServerInfo() (server interface{}, err error) {
	var data utils.Server
	data.Os = utils.InitOS()
	if data.Cpu, err = utils.InitCPU(); err != nil {
		s.serviceCtx.Logger.Error("func utils.InitCPU() Failed", zap.String("err", err.Error()))
		return &data, err
	}
	if data.Ram, err = utils.InitRAM(); err != nil {
		s.serviceCtx.Logger.Error("func utils.InitRAM() Failed", zap.String("err", err.Error()))
		return &data, err
	}
	if data.Disk, err = utils.InitDisk([]string{"/"}); err != nil {
		s.serviceCtx.Logger.Error("func utils.InitDisk() Failed", zap.String("err", err.Error()))
		return &data, err
	}

	return &data, nil
}
