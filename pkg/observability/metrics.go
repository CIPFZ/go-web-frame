package observability

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
)

// Package observability -----------------------------
// @file        : metrics.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:49
// @description :
// -------------------------------------------

// InitMeter 初始化指标器
func InitMeter(cfg Config) (*metric.MeterProvider, error) {
	exporter, err := CreateMetricExporter(cfg)
	if err != nil {
		return nil, err
	}

	res, err := CreateResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}
