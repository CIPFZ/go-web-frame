package system

import (
	"github.com/CIPFZ/gowebframe/internal/api/v1/system"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
)

type SysRouter struct {
	serviceCtx *svc.ServiceContext
}

func (s *SysRouter) InitSystemRouter(Router *gin.RouterGroup) {
	systemApi := system.NewSysApi(s.serviceCtx)
	sysRouter := Router.Group("system")
	sysRouter.POST("getSystemConfig", systemApi.GetSystemConfig) // 获取配置文件内容
	sysRouter.POST("getServerInfo", systemApi.GetServerInfo)     // 获取服务器信息
}
