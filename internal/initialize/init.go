package initialize

import (
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
)

// SetupHandlers 初始化全局函数
func SetupHandlers(serviceCtx *svc.ServiceContext) {
	// 注册系统重载处理函数
	utils.GlobalSystemEvents.RegisterReloadHandler(func() error {
		return nil
	})
}
