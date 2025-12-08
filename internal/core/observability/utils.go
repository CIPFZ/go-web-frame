package observability

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// createResource 创建 otel resource
func createResource(cfg config.Observability) (*resource.Resource, error) {
	return resource.New(context.Background(),
		resource.WithAttributes(
			semconv.HTTPScheme(semconv.SchemaURL),
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
}

// StandardizedEndpoint ✨ (辅助函数) OTLP gRPC 规范要求 http:// 或 https:// 协议头
// 我们的配置 "localhost:4317" 不符合规范，这个函数进行修正 (注意: OTLP/HTTP 导出器会自动处理, 但 gRPC 不会)
func StandardizedEndpoint(cfg config.OtelExporterConfig) string {
	if cfg.Protocol == "grpc" {
		// gRPC 导出器现在不需要 http/https 前缀了
		// 并且 otlploggrpc, otlptracegrpc, otlpmetricgrpc 行为统一
		return cfg.Endpoint
	}

	// HTTP 导出器则需要
	//if cfg.Protocol == "http" {
	//	if !strings.HasPrefix(cfg.Endpoint, "http") {
	//		// 假设，如果不安全则为 http
	//		if cfg.Insecure {
	//			return "http://" + cfg.Endpoint
	//		}
	//		return "https://" + cfg.Endpoint
	//	}
	//}
	return cfg.Endpoint
}
