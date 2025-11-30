package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CIPFZ/gowebframe/internal/initialize"
	"github.com/CIPFZ/gowebframe/internal/server"
	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/pkg/i18n"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/CIPFZ/gowebframe/pkg/observability"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	sdklog "go.opentelemetry.io/otel/sdk/log"

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

	// initializeSystem 现在返回一个聚合的关停函数列表
	allShutdowns, err := initializeSystem(*configPath, serviceCtx)
	if err != nil {
		// 如果初始化失败 (e.g., GORM panic), 我们在这里 log.Fatal
		log.Fatalf("系统初始化失败: %v", err)
	}
	// ✨ 关键：defer Sync 必须在 main 函数中
	defer logger.Sync(serviceCtx.Logger)

	// ---------------- 2. 启动 HTTP 服务 (在 Goroutine 中) ----------------
	httpServer := server.NewServer(serviceCtx)

	// 创建一个 channel 来接收 ListenAndServe 的错误
	serverErr := make(chan error, 1)
	go func() {
		serviceCtx.Logger.Info("HTTP服务启动中...", zap.String("addr", httpServer.Addr))
		// 打印banner信息
		server.PrintBanner(httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err // 将启动错误发送到 main
		}
	}()

	// ---------------- 3. 阻塞并等待退出 ----------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		// 服务启动失败 (例如，端口被占用)
		serviceCtx.Logger.Fatal("HTTP服务启动失败", zap.Error(err))
	case s := <-quit:
		// 收到操作系统退出信号
		serviceCtx.Logger.Info("收到退出信号，准备关闭服务...", zap.String("signal", s.String()))
	}

	// ---------------- 4. 优雅退出 ----------------
	// 创建一个总的关停超时 Context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 关停 HTTP 服务 (它现在只返回 error, 不再 panic)
	if err := server.ShutdownServer(ctx, serviceCtx); err != nil {
		serviceCtx.Logger.Warn("HTTP服务关闭异常", zap.Error(err))
	}

	// 2. 关键：按顺序执行所有其他的关停 (Otel, Mongo...)
	for _, shutdown := range allShutdowns {
		if err := shutdown(ctx); err != nil {
			serviceCtx.Logger.Warn("组件关停异常", zap.Error(err))
		}
	}

	// 3. 最后关停 Logger
	logger.Sync(serviceCtx.Logger)

	serviceCtx.Logger.Info("服务已成功关闭")
}

// initializeSystem 初始化系统所有组件
func initializeSystem(configPath string, serviceCtx *svc.ServiceContext) ([]func(context.Context) error, error) {
	var shutdowns []func(context.Context) error
	// Step1: 加载配置
	serviceCtx.Viper = utils.MustInit(func() (*viper.Viper, error) {
		return initialize.InitConfig(configPath, serviceCtx)
	}, "Config")

	initialize.OtherInit(serviceCtx)

	// Step 2: 初始化 Otel (三大支柱) 在 Logger 和 数据库 之前
	cfg := serviceCtx.Config
	var lp *sdklog.LoggerProvider
	var logShutdown observability.ShutdownFunc
	{
		var err error
		lp, logShutdown, err = observability.InitLogs(cfg.Observable)
		if err != nil {
			return nil, fmt.Errorf("failed to init logs: %w", err)
		}

		var traceShutdown observability.ShutdownFunc
		traceShutdown, err = observability.InitTraces(cfg.Observable)
		if err != nil {
			return nil, fmt.Errorf("failed to init traces: %w", err)
		}
		shutdowns = append(shutdowns, traceShutdown)

		var metricShutdown observability.ShutdownFunc
		metricShutdown, err = observability.InitMetrics(cfg.Observable)
		if err != nil {
			return nil, fmt.Errorf("failed to init metrics: %w", err)
		}
		fmt.Println("正在启动 Go Runtime 指标收集...")
		if err = runtime.Start(
			runtime.WithMeterProvider(otel.GetMeterProvider()),      // 使用我们全局的 MeterProvider
			runtime.WithMinimumReadMemStatsInterval(15*time.Second), // 收集间隔 (默认15秒 )
		); err != nil {
			return nil, fmt.Errorf("failed to start runtime metrics: %w", err)
		}
		shutdowns = append(shutdowns, metricShutdown)
	}

	// Step 3: 初始化 Logger (打磨后的)
	serviceCtx.Logger = utils.MustInit(func() (*zap.Logger, error) {
		// (我们打磨后的 OTELLoggerConfig 只需要 Provider)
		otelLogCfg := &logger.OTELLoggerConfig{
			LogProvider: lp,
		}
		return logger.NewLogger(&cfg.Logger, otelLogCfg)
	}, "Logger")
	zap.ReplaceGlobals(serviceCtx.Logger)
	serviceCtx.Logger.Info("Logger 初始化成功 (已桥接 Otel)")

	// Step3: 初始化 I18n
	serviceCtx.I18n = utils.MustInit(func() (*i18n.Service, error) {
		return i18n.NewI18n(&serviceCtx.Config.I18n, serviceCtx.Logger)
	}, "I18n")

	// Step 5: 数据库连接 (已打磨, 包含 panic 和 otel)
	serviceCtx.DB = initialize.Gorm(serviceCtx)
	initialize.MustInitMongo(serviceCtx)
	initialize.MustInitRedis(serviceCtx)

	// 将 Mongo 关停添加到列表中
	if serviceCtx.Mongo != nil {
		shutdowns = append(shutdowns, func(ctx context.Context) error {
			// ShutdownMongo 内部会创建自己的 5s 超时
			initialize.ShutdownMongo(serviceCtx)
			return nil
		})
	}
	// 注册权限鉴权
	serviceCtx.CasbinEnforcer = utils.InitCasbin(serviceCtx.DB)
	serviceCtx.Logger.Info("Casbin 初始化完成")

	// Step6: 初始化定时任务
	initialize.Timer(serviceCtx)
	// Step7: 注册全局函数
	initialize.SetupHandlers(serviceCtx)
	// Step8: 添加 JWT
	serviceCtx.JWT = utils.NewJWT(serviceCtx.Config.JWT, serviceCtx.Logger, serviceCtx.Redis)
	// Step 9: 初始化操作日志服务
	opLogService := systemService.NewOperationLogService(serviceCtx.DB, serviceCtx.Logger)
	serviceCtx.OperationLogService = opLogService
	shutdowns = append(shutdowns, opLogService.Close)

	// Step 10: 将 Otel Log 关停函数最后添加到列表
	shutdowns = append(shutdowns, logShutdown)

	serviceCtx.Logger.Info("✅ 系统初始化成功")
	return shutdowns, nil
}
