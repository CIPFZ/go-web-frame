package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/CIPFZ/gowebframe/internal/initialize"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"go.uber.org/zap"
)

// NewServer 创建 HTTP 服务
func NewServer(serviceCtx *svc.ServiceContext) *http.Server {
	// 初始化路由
	engine := initialize.InitRouters(serviceCtx)
	address := fmt.Sprintf(":%d", serviceCtx.Config.System.Port)

	serviceCtx.SRV = &http.Server{
		Addr:           address,
		Handler:        engine,
		ReadTimeout:    10 * time.Minute,
		WriteTimeout:   10 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}
	return serviceCtx.SRV
}

// ShutdownServer 优雅关闭 HTTP 服务
func ShutdownServer(ctx context.Context, serviceCtx *svc.ServiceContext) error {
	serviceCtx.Logger.Info("正在关闭 HTTP 服务...")
	if err := serviceCtx.SRV.Shutdown(ctx); err != nil {
		serviceCtx.Logger.Error("HTTP服务关闭异常", zap.Error(err))
		return err
	}
	serviceCtx.Logger.Info("HTTP 服务已优雅退出 ✅")
	return nil
}
