package observability

import (
	"context"
	"log"
	"strings"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Package observability -----------------------------
// @file        : config.go
// @author      : CIPFZ
// @time        : 2025/9/23 14:37
// @description :
// -------------------------------------------

type Config struct {
	ServiceName      string             `yaml:"service_name" json:"service_name" toml:"service_name" mapstructure:"service_name"`
	ServiceVersion   string             `yaml:"service_version" json:"service_version" toml:"service_version" mapstructure:"service_version"`
	Exporter         string             `yaml:"exporter" json:"exporter" toml:"exporter" mapstructure:"exporter"`
	Environment      string             `yaml:"environment" json:"environment" toml:"environment" mapstructure:"environment"`
	TraceSampleRatio float64            `yaml:"trace_sample_ratio" json:"trace_sample_ratio" toml:"trace_sample_ratio" mapstructure:"trace_sample_ratio"`
	OtelExporter     OtelExporterConfig `yaml:"otel_exporter" json:"otel_exporter" toml:"otel_exporter" mapstructure:"otel_exporter"`
}

// OtelExporterConfig 配置
type OtelExporterConfig struct {
	Protocol      string `yaml:"protocol" json:"protocol" toml:"protocol" mapstructure:"protocol"`
	Endpoint      string `yaml:"endpoint" json:"endpoint" toml:"endpoint" mapstructure:"endpoint"`
	Authorization string `yaml:"authorization" json:"authorization" toml:"authorization" mapstructure:"authorization"`
	Organization  string `yaml:"organization" json:"organization" toml:"organization" mapstructure:"organization"`
	StreamName    string `yaml:"stream_name" json:"stream_name" toml:"stream_name" mapstructure:"stream_name"`
	Insecure      bool   `yaml:"insecure" json:"insecure" toml:"insecure" mapstructure:"insecure"`
}

// createResource 创建 otel resource
func createResource(cfg Config) (*resource.Resource, error) {
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
func (cfg OtelExporterConfig) StandardizedEndpoint() string {
	if cfg.Protocol == "grpc" {
		// gRPC 导出器现在不需要 http/https 前缀了
		// 并且 otlploggrpc, otlptracegrpc, otlpmetricgrpc 行为统一
		return cfg.Endpoint
	}

	// HTTP 导出器则需要
	if cfg.Protocol == "http" {
		if !strings.HasPrefix(cfg.Endpoint, "http") {
			// 假设，如果不安全则为 http
			if cfg.Insecure {
				return "http://" + cfg.Endpoint
			}
			return "https://" + cfg.Endpoint
		}
	}
	return cfg.Endpoint
}

// ShutdownFunc 定义了一个统一的关闭函数签名
type ShutdownFunc func(context.Context) error

// 定义一个通用接口：所有 OTel Provider 都实现 Shutdown(context.Context) error
type shutdowner interface {
	Shutdown(ctx context.Context) error
}

// SafeShutdown 封装 Shutdown，保证 nil 安全，错误打印
func SafeShutdown(p shutdowner) {
	if p == nil {
		return
	}
	if err := p.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown provider: %v", err)
	}
}
