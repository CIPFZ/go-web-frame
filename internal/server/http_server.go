package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CIPFZ/gowebframe/internal/initialize"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"go.uber.org/zap"
)

// RunServer 启动 HTTP 服务
func RunServer(serviceCtx *svc.ServiceContext) {
	// 初始化路由
	engine := initialize.InitRouters(serviceCtx)
	address := fmt.Sprintf(":%d", serviceCtx.Config.System.Port)
	// 打印 banner
	printBanner(address)

	serviceCtx.SRV = &http.Server{
		Addr:           address,
		Handler:        engine,
		ReadTimeout:    10 * time.Minute,
		WriteTimeout:   10 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := serviceCtx.SRV.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serviceCtx.Logger.Fatal("HTTP服务启动失败", zap.Error(err))
		}
	}()

	serviceCtx.Logger.Info("HTTP服务启动成功", zap.String("addr", address))
}

// ShutdownServer 优雅关闭 HTTP 服务
func ShutdownServer(serviceCtx *svc.ServiceContext) {
	serviceCtx.Logger.Info("正在关闭 HTTP 服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serviceCtx.SRV.Shutdown(ctx); err != nil {
		serviceCtx.Logger.Fatal("HTTP服务关闭异常", zap.Error(err))
	}
	serviceCtx.Logger.Info("HTTP 服务已优雅退出 ✅")
}
