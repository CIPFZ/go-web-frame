package observability

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Package observability -----------------------------
// @file        : tracing.go
// @author      : CIPFZ
// @time        : 2025/9/19 18:27
// @description :
// -------------------------------------------

// InitTracer 初始化 Tracer Provider
func InitTracer(cfg Config) (*sdktrace.TracerProvider, error) {
	exporter, err := CreateTraceExporter(cfg)
	if err != nil {
		return nil, err
	}

	// 定义 Resource (服务名、版本等)
	res, err := CreateResource(cfg)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.Otel.TraceSampleRatio)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}
