package main

import (
	"flag"
	"fmt"
	"go.opentelemetry.io/otel"

	"github.com/CIPFZ/gowebframe/internal/api"
	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/internal/global"
	"github.com/CIPFZ/gowebframe/pkg/i18n"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/CIPFZ/gowebframe/pkg/observability"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

// Package server -----------------------------
// @file        : main.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:21
// @description :
// -------------------------------------------

func main() { // 命令行参数 -f 指定配置文件路径
	configPath := flag.String("f", "etc/config.yaml", "config file path")
	flag.Parse()

	// Step1: 加载配置
	cfg, err := config.InitConfig(*configPath)
	if err != nil {
		panic(err)
	}
	global.G.Config = cfg

	// Step2: 初始化 Logger + 配置日志推送
	lp, err := observability.InitLogs(cfg.Observable)
	if err != nil {
		fmt.Printf("failed to init logs error: %+v\n", err)
	}
	defer observability.SafeShutdown(lp)

	log, err := logger.NewLogger(&cfg.Logger, &logger.OTELLoggerConfig{
		LogProvider: lp,
		ServiceName: cfg.Observable.ServiceName,
		ServiceVer:  cfg.Observable.ServiceVersion,
		Environment: cfg.App.Environment,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Sync(log)
	global.G.Logger = log
	// 全局替换 logger - zap.L()
	zap.ReplaceGlobals(log)

	// Step3: 初始化 I18n
	i18nService, err := i18n.NewI18n(&cfg.I18n, log)
	if err != nil {
		global.G.Logger.Fatal("failed to init i18n", zap.Error(err))
	}
	global.G.I18n = i18nService

	// Step4: 初始化链路追踪
	tp, err := observability.InitTracer(cfg.Observable)
	if err != nil {
		global.G.Logger.Fatal("failed init tracer", zap.Error(err))
	}
	defer observability.SafeShutdown(tp)

	// Step5: 开启指标meter
	mp, err := observability.InitMeter(cfg.Observable)
	if err != nil {
		global.G.Logger.Fatal("failed init meter", zap.Error(err))
	}
	defer observability.SafeShutdown(mp)

	// 创建 HTTP Metrics
	meter := otel.Meter(cfg.Observable.ServiceName)
	httpMetrics, _ := observability.NewHTTPMetrics(meter)

	// 初始化 Gin
	r := gin.Default()
	r.Use(otelgin.Middleware(cfg.Observable.ServiceName)) // tracing
	r.Use(httpMetrics.Middleware())                       // metrics
	r.Use(logger.GinLoggerMiddleware(log))                // log

	// 注册路由
	api.RegisterRoutes(r)

	// 启动服务
	if err := r.Run(":8080"); err != nil {
		global.G.Logger.Fatal("failed to start server", zap.Error(err))
	}
}
