package observability

import (
	"fmt"

	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// Package observability -----------------------------
// @file        : logs.go
// @author      : CIPFZ
// @time        : 2025/9/23 21:54
// @description :
// -------------------------------------------

// InitLogs 初始化日志推送 provider
func InitLogs(cfg Config) (*sdklog.LoggerProvider, error) {
	exporter, err := CreateLogsExporter(cfg)
	if err != nil {
		return nil, err
	}

	res, err := CreateResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	return lp, nil
}
