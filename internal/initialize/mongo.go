package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/CIPFZ/gowebframe/internal/svc"
	"go.opentelemetry.io/otel"

	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/event"
	opt "go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.uber.org/zap"
)

// MustInitMongo 初始化 MongoDB 连接
func MustInitMongo(serviceCtx *svc.ServiceContext) {
	if !serviceCtx.Config.System.UseMongo {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := serviceCtx.Config.Mongo
	config := &qmgo.Config{
		Uri:              cfg.Uri(),
		Database:         cfg.Database,
		MinPoolSize:      &cfg.MinPoolSize,
		MaxPoolSize:      &cfg.MaxPoolSize,
		SocketTimeoutMS:  &cfg.SocketTimeoutMs,
		ConnectTimeoutMS: &cfg.ConnectTimeoutMs,
	}

	if cfg.Username != "" && cfg.Password != "" {
		config.Auth = &qmgo.Credential{
			Username:   cfg.Username,
			Password:   cfg.Password,
			AuthSource: cfg.AuthSource,
		}
	}

	var opts []options.ClientOptions
	// Otel Monitor 会自动捕获命令、耗时、错误并创建 Span
	monitor := otelmongo.NewMonitor(
		otelmongo.WithTracerProvider(otel.GetTracerProvider()),
	)
	if cfg.IsZap {
		opts = append(opts, zapOptions()) // 自定义 zap logger
	}
	opts = append(opts, options.ClientOptions{
		ClientOptions: &opt.ClientOptions{
			Monitor: monitor,
		},
	})

	c, err := qmgo.Open(ctx, config, opts...)
	if err != nil {
		panic(fmt.Errorf("init mongo failed: %w", err))
	}

	// 添加 Ping 健康检查 (保留 panic)
	if err := c.Ping(int64(5 * time.Second)); err != nil {
		panic(fmt.Errorf("mongo ping failed: %w", err))
	}

	serviceCtx.Mongo = c
	serviceCtx.Logger.Info("✅ MongoDB 连接成功 (已启用 Otel 追踪)")
	err = initIndexes(ctx)
	if err != nil {
		panic(err)
	}
}

// ShutdownMongo 优雅关闭 Mongo
func ShutdownMongo(serviceCtx *svc.ServiceContext) {
	if serviceCtx.Mongo != nil {
		// ✨ 2. 使用 5 秒超时，而不是 Background()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		serviceCtx.Logger.Info("🛑 正在关闭 MongoDB...")
		if err := serviceCtx.Mongo.Close(ctx); err != nil {
			serviceCtx.Logger.Error("MongoDB 关闭异常", zap.Error(err))
			return
		}
		serviceCtx.Logger.Info("MongoDB 已关闭")
	}
}

// zapOptions 返回用于 MongoDB 的 zap 监控 options
func zapOptions() options.ClientOptions {
	cmdMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, e *event.CommandStartedEvent) {
			zap.L().Info(
				fmt.Sprintf("[MongoDB][RequestID:%d][DB:%s] %s", e.RequestID, e.DatabaseName, e.Command),
				zap.String("business", "mongo"),
			)
		},
		Succeeded: func(ctx context.Context, e *event.CommandSucceededEvent) {
			zap.L().Info(
				fmt.Sprintf("[MongoDB][RequestID:%d][%s] %s", e.RequestID, e.Duration, e.Reply),
				zap.String("business", "mongo"),
			)
		},
		Failed: func(ctx context.Context, e *event.CommandFailedEvent) {
			zap.L().Error(
				fmt.Sprintf("[MongoDB][RequestID:%d][%s] %s", e.RequestID, e.Duration, e.Failure),
				zap.String("business", "mongo"),
			)
		},
	}

	return options.ClientOptions{
		ClientOptions: &opt.ClientOptions{
			Monitor: cmdMonitor,
		},
	}
}

// 初始化项目所需索引
func initIndexes(ctx context.Context) error {
	indexConfig := map[string][][]string{}

	for collection, indexes := range indexConfig {
		if err := createIndexes(ctx, collection, indexes); err != nil {
			return err
		}
	}
	return nil
}

func createIndexes(ctx context.Context, collName string, indexSets [][]string) error {
	// TODO
	return nil
}
