package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	"github.com/CIPFZ/gowebframe/internal/core/file"
	"github.com/CIPFZ/gowebframe/internal/core/jwt"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"go.opentelemetry.io/otel"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/audit"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/i18n"
	corelog "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/core/observability"
	"github.com/CIPFZ/gowebframe/internal/core/server"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
)

func main() {
	// ---------------- 1. 初始化系统 ----------------
	configPath := flag.String("f", "C:\\Users\\ytq\\work\\go\\go-web-frame\\configs\\config.yaml", "config file path")
	flag.Parse()

	serviceCtx := svc.NewServiceContext()

	// 初始化并获取关停函数列表
	allShutdowns, err := initializeSystem(*configPath, serviceCtx)
	if err != nil {
		log.Fatalf("❌ 系统初始化失败: %v", err)
	}

	// ✨ 确保 Logger 最后刷盘 (stdout/file)
	defer corelog.Sync(serviceCtx.Logger)

	// ---------------- 2. 启动 HTTP 服务 ----------------
	// server.NewServer 内部会调用 InitRouters 进行依赖注入和路由注册
	serviceCtx.SRV = server.NewServer(serviceCtx)
	server.PrintBanner(serviceCtx.Config.System.Port)

	serverErr := make(chan error, 1)
	go func() {
		serviceCtx.Logger.Info("🚀 HTTP服务启动中...", zap.String("addr", serviceCtx.SRV.Addr))
		if listenErr := serviceCtx.SRV.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			serverErr <- listenErr
		}
	}()

	// ---------------- 3. 阻塞并等待退出 ----------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		serviceCtx.Logger.Fatal("❌ HTTP服务启动失败", zap.Error(err))
	case s := <-quit:
		serviceCtx.Logger.Info("🛑 收到退出信号，准备关闭服务...", zap.String("signal", s.String()))
	}

	// ---------------- 4. 优雅退出 ----------------
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// 1. 关停 HTTP 服务 (停止接收新请求)
	if shutdownErr := server.ShutdownServer(shutdownCtx, serviceCtx); shutdownErr != nil {
		serviceCtx.Logger.Warn("HTTP服务关闭异常", zap.Error(shutdownErr))
	}

	// 2. 执行组件关停 (DB, Mongo, Redis, Otel Traces/Metrics, Audit)
	// 按照 append 的顺序执行
	for _, shutdown := range allShutdowns {
		if err := shutdown(shutdownCtx); err != nil {
			serviceCtx.Logger.Warn("组件关停异常", zap.Error(err))
		}
	}

	// (defer logger.Sync 会在这里执行)
	log.Println("👋 服务已成功关闭")
}

// initializeSystem 初始化核心组件并组装
func initializeSystem(path string, serviceCtx *svc.ServiceContext) ([]utils.ShutdownFunc, error) {
	var shutdowns []utils.ShutdownFunc
	var lp *sdklog.LoggerProvider
	var logShutdown utils.ShutdownFunc
	var err error

	// Step 1: 加载配置 (使用 core/config)
	var v *viper.Viper
	cfg, v, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("config load failed: %w", err)
	}
	serviceCtx.Config = cfg
	serviceCtx.Viper = v

	// Step 2: 初始化 Otel (core/trace)
	{
		// Logs
		lp, logShutdown, err = observability.InitLogs(cfg.Observable)
		if err != nil {
			return nil, fmt.Errorf("otel logs init failed: %w", err)
		}

		// Traces
		traceShutdown, err := observability.InitTraces(cfg.Observable)
		if err != nil {
			return nil, fmt.Errorf("otel traces init failed: %w", err)
		}
		shutdowns = append(shutdowns, traceShutdown)

		// Metrics
		metricShutdown, err := observability.InitMetrics(cfg.Observable)
		if err != nil {
			return nil, fmt.Errorf("otel metrics init failed: %w", err)
		}
		shutdowns = append(shutdowns, metricShutdown)

		// Runtime Metrics
		fmt.Println("正在启动 Go Runtime 指标收集...")
		if err = runtime.Start(
			runtime.WithMeterProvider(otel.GetMeterProvider()),
			runtime.WithMinimumReadMemStatsInterval(15*time.Second),
		); err != nil {
			return nil, fmt.Errorf("runtime metrics start failed: %w", err)
		}
	}

	// Step 3: 初始化 Logger (core/log)
	serviceCtx.Logger, err = corelog.NewLogger(&cfg.Logger, &config.OTELLoggerConfig{
		LogProvider: lp, // 传入 SDK Provider
	})
	if err != nil {
		return nil, fmt.Errorf("logger init failed: %w", err)
	}
	// 替换全局 Logger，方便 middleware 使用 zap.L()
	zap.ReplaceGlobals(serviceCtx.Logger)
	// 修正: global.SetLoggerProvider 已经在 InitLogs 里做了，这里不需要再做

	serviceCtx.Logger.Info("✅ 基础设施初始化：配置、可观测性、日志")

	// Step 4: 国际化 (core/i18n)
	serviceCtx.I18n, err = i18n.NewI18n(cfg.I18n, serviceCtx.Logger)
	if err != nil {
		return nil, fmt.Errorf("i18n init failed: %w", err)
	}

	// Step 5: 数据库连接 (core/db)
	// MySQL (GORM)
	serviceCtx.DB, err = db.InitMysql(cfg.Mysql, serviceCtx.Logger)
	if err != nil {
		return nil, fmt.Errorf("mysql init failed: %w", err)
	}

	// Redis
	serviceCtx.Redis, err = db.InitRedis(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis init failed: %w", err)
	}

	// Mongo
	if cfg.System.UseMongo {
		serviceCtx.Mongo, err = db.InitMongo(cfg.Mongo)
		if err != nil {
			return nil, fmt.Errorf("mongo init failed: %w", err)
		}
		// 注册 Mongo 关停
		shutdowns = append(shutdowns, func(ctx context.Context) error {
			serviceCtx.Logger.Info("🛑 正在关闭 MongoDB...")
			return serviceCtx.Mongo.Close(ctx)
		})
	}

	// Step 6: 权限 Casbin
	serviceCtx.CasbinEnforcer = claims.InitCasbin(serviceCtx.DB)
	serviceCtx.Logger.Info("Casbin 初始化完成")

	// Step 7: JWT (pkg/utils)
	// 依赖 Redis (ServiceContext已持有)
	serviceCtx.JWT = jwt.NewJWT(serviceCtx.Config.JWT, serviceCtx.Logger, serviceCtx.Redis)

	// Step 8: 审计日志 (core/audit)
	// ✨ 关键：使用 core 层的 AuditRecorder，解耦循环依赖
	auditRecorder := audit.NewAuditRecorder(serviceCtx.DB, serviceCtx.Logger)
	serviceCtx.AuditRecorder = auditRecorder
	shutdowns = append(shutdowns, auditRecorder.Close)

	// Step 9: 初始化 OSS
	serviceCtx.OSS = file.NewFileService(serviceCtx.Config.File, serviceCtx.Logger)

	// Step 10: 最后添加 Otel Log Flush (确保它最后执行)
	shutdowns = append(shutdowns, logShutdown)

	serviceCtx.Logger.Info("✅ 系统核心组件组装完成")
	return shutdowns, nil
}
