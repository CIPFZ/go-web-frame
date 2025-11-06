package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	"github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SysApi struct {
	service system.IBaseService
}

func NewSysApi(serviceCtx *svc.ServiceContext) *SysApi {
	return &SysApi{service: system.NewBaseService(serviceCtx)}
}

// GetSystemConfig 获取配置文件内容
func (s *SysApi) GetSystemConfig(c *gin.Context) {
	log := logger.GetLogger(c)
	config, err := s.service.GetSystemConfig()
	if err != nil {
		log.Error("获取系统配置失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(systemRes.SysConfigResponse{Config: config}, "获取成功", c)
}

// GetServerInfo 获取服务器信息
func (s *SysApi) GetServerInfo(c *gin.Context) {
	log := logger.GetLogger(c)
	server, err := s.service.GetServerInfo()
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"server": server}, "获取成功", c)
}
