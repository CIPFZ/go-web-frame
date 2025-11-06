package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CIPFZ/gowebframe/internal/initialize"
	"github.com/CIPFZ/gowebframe/internal/server"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/pkg/i18n"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/CIPFZ/gowebframe/pkg/observability"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Package server -----------------------------
// @file        : main.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:21
// @description :
// -------------------------------------------

func main() {
	// ---------------- 1. 初始化系统 ----------------
	configPath := flag.String("f", "etc/config.yaml", "config file path")
	flag.Parse()

	serviceCtx := svc.NewServiceContext()
	initializeSystem(*configPath, serviceCtx)

	// ---------------- 2. 启动 HTTP 服务 ----------------
	server.RunServer(serviceCtx)

	// ---------------- 3. 优雅退出 ----------------
	waitForExitSignal(serviceCtx)
	server.ShutdownServer(serviceCtx)
	initialize.ShutdownMongo(serviceCtx)
}

// initializeSystem 初始化系统所有组件
func initializeSystem(configPath string, serviceCtx *svc.ServiceContext) {
	// Step1: 加载配置
	serviceCtx.Viper = utils.MustInit(func() (*viper.Viper, error) {
		return initialize.InitConfig(configPath, serviceCtx)
	}, "Config")

	initialize.OtherInit(serviceCtx)

	// Step2: 初始化 Logger + 配置日志推送
	lp, err := observability.InitLogs(serviceCtx.Config.Observable)
	if err != nil {
		fmt.Printf("failed to init logs error: %+v\n", err)
	}
	defer observability.SafeShutdown(lp)

	serviceCtx.Logger = utils.MustInit(func() (*zap.Logger, error) {
		return logger.NewLogger(&serviceCtx.Config.Logger, &logger.OTELLoggerConfig{
			LogProvider: lp,
			ServiceName: serviceCtx.Config.Observable.ServiceName,
			ServiceVer:  serviceCtx.Config.Observable.ServiceVersion,
			Environment: serviceCtx.Config.System.Environment,
		})
	}, "Logger")
	defer logger.Sync(serviceCtx.Logger)
	// 全局替换 logger - zap.L()
	zap.ReplaceGlobals(serviceCtx.Logger)

	// Step3: 初始化 I18n
	serviceCtx.I18n = utils.MustInit(func() (*i18n.Service, error) {
		return i18n.NewI18n(&serviceCtx.Config.I18n, serviceCtx.Logger)
	}, "I18n")

	// Step5: 数据库连接
	serviceCtx.DB = initialize.Gorm(serviceCtx)
	// Mongo 连接
	initialize.MustInitMongo(serviceCtx)
	// Redis 连接
	initialize.MustInitRedis(serviceCtx)
	// Step6: 初始化定时任务
	initialize.Timer(serviceCtx)
	// Step6: 注册全局函数
	initialize.SetupHandlers(serviceCtx)
	// Step8: 初始化监控指标
	initialize.InitObservability(serviceCtx)
}

// 等待退出信号
func waitForExitSignal(serviceCtx *svc.ServiceContext) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	serviceCtx.Logger.Info("收到退出信号，准备关闭服务...", zap.String("signal", s.String()))
}
