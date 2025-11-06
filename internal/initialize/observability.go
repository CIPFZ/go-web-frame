package initialize

import (
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/observability"
	"go.uber.org/zap"
)

func InitObservability(serviceCtx *svc.ServiceContext) {
	cfg := serviceCtx.Config.Observable
	// TODO 对 Observable 进行校验
	// 初始化链路追踪
	tp, err := observability.InitTracer(cfg)
	if err != nil {
		serviceCtx.Logger.Fatal("failed init tracer", zap.Error(err))
	}
	defer observability.SafeShutdown(tp)

	// 开启指标meter
	mp, err := observability.InitMeter(cfg)
	if err != nil {
		serviceCtx.Logger.Fatal("failed init meter", zap.Error(err))
	}
	defer observability.SafeShutdown(mp)

	// 创建 HTTP Metrics
	//meter := otel.Meter(serviceCtx.Config.Observable.ServiceName)
}
