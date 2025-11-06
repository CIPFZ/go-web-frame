package observability

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Package observability -----------------------------
// @file        : config.go
// @author      : CIPFZ
// @time        : 2025/9/23 14:37
// @description :
// -------------------------------------------

type Config struct {
	ServiceName    string `yaml:"service_name" json:"service_name" toml:"service_name" mapstructure:"service_name"`
	ServiceVersion string `yaml:"service_version" json:"service_version" toml:"service_version" mapstructure:"service_version"`
	Exporter       string `yaml:"exporter" json:"exporter" toml:"exporter" mapstructure:"exporter"`
	Otel           Otel   `yaml:"otel" json:"otel" toml:"otel" mapstructure:"otel"`
}

// Otel 配置
type Otel struct {
	Protocol         string  `yaml:"protocol" json:"protocol" toml:"protocol" mapstructure:"protocol"`
	Endpoint         string  `yaml:"endpoint" json:"endpoint" toml:"endpoint" mapstructure:"endpoint"`
	Authorization    string  `yaml:"authorization" json:"authorization" toml:"authorization" mapstructure:"authorization"`
	Organization     string  `yaml:"organization" json:"organization" toml:"organization" mapstructure:"organization"`
	StreamName       string  `yaml:"stream_name" json:"stream_name" toml:"stream_name" mapstructure:"stream_name"`
	Insecure         bool    `yaml:"insecure" json:"insecure" toml:"insecure" mapstructure:"insecure"`
	TraceSampleRatio float64 `yaml:"trace_sample_ratio" json:"trace_sample_ratio" toml:"trace_sample_ratio" mapstructure:"trace_sample_ratio"`
}

// CreateTraceExporter 创建Trace导出器
func CreateTraceExporter(cfg Config) (sdktrace.SpanExporter, error) {
	var exporter sdktrace.SpanExporter
	var err error

	headers := map[string]string{
		"Authorization": cfg.Otel.Authorization,
		"organization":  cfg.Otel.Organization,
		"stream-name":   cfg.Otel.StreamName,
	}

	switch cfg.Exporter {
	case "otel":
		switch cfg.Otel.Protocol {
		case "grpc":
			opts := []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint(cfg.Otel.Endpoint),
				otlptracegrpc.WithHeaders(headers),
			}
			if cfg.Otel.Insecure {
				opts = append(opts, otlptracegrpc.WithInsecure())
			} else {
				opts = append(opts, otlptracegrpc.WithDialOption(
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				))
			}
			exporter, err = otlptracegrpc.New(context.Background(), opts...)

		case "http":
			opts := []otlptracehttp.Option{
				otlptracehttp.WithEndpoint(cfg.Otel.Endpoint),
				otlptracehttp.WithURLPath(fmt.Sprintf("/api/%s/v1/traces", cfg.Otel.Organization)),
				otlptracehttp.WithHeaders(headers),
				otlptracehttp.WithInsecure(),
			}
			exporter, err = otlptracehttp.New(context.Background(), opts...)

		default:
			return nil, fmt.Errorf("unsupported otel protocol: %s", cfg.Otel.Protocol)
		}

	default:
		// fallback: stdout
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	return exporter, nil
}

// CreateMetricExporter 创建Metric导出器
func CreateMetricExporter(cfg Config) (metric.Exporter, error) {
	var exporter metric.Exporter
	var err error

	headers := map[string]string{
		"Authorization": cfg.Otel.Authorization,
		"organization":  cfg.Otel.Organization,
		"stream-name":   cfg.Otel.StreamName,
	}

	switch cfg.Exporter {
	case "otel":
		switch cfg.Otel.Protocol {
		case "grpc":
			opts := []otlpmetricgrpc.Option{
				otlpmetricgrpc.WithEndpoint(cfg.Otel.Endpoint),
				otlpmetricgrpc.WithHeaders(headers),
			}
			if cfg.Otel.Insecure {
				opts = append(opts, otlpmetricgrpc.WithInsecure())
			} else {
				opts = append(opts, otlpmetricgrpc.WithDialOption(
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				))
			}
			exporter, err = otlpmetricgrpc.New(context.Background(), opts...)

		case "http":
			opts := []otlpmetrichttp.Option{
				otlpmetrichttp.WithEndpoint(cfg.Otel.Endpoint),
				otlpmetrichttp.WithURLPath(fmt.Sprintf("/api/%s/v1/metrics", cfg.Otel.Organization)),
				otlpmetrichttp.WithHeaders(headers),
				otlpmetrichttp.WithInsecure(),
			}
			exporter, err = otlpmetrichttp.New(context.Background(), opts...)

		default:
			return nil, fmt.Errorf("unsupported otel protocol: %s", cfg.Otel.Protocol)
		}

	default:
		// fallback: stdout
		exporter, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}
	return exporter, nil
}

// CreateLogsExporter 创建 Logs 导出器
func CreateLogsExporter(cfg Config) (sdklog.Exporter, error) {
	if cfg.Otel.Protocol == "" || cfg.Otel.Endpoint == "" {
		return nil, nil
	}
	var exp sdklog.Exporter
	var err error

	headers := map[string]string{
		"Authorization": cfg.Otel.Authorization,
		"organization":  cfg.Otel.Organization,
		"stream-name":   cfg.Otel.StreamName,
	}

	switch cfg.Otel.Protocol {
	case "grpc":
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(cfg.Otel.Endpoint),
			otlploggrpc.WithHeaders(headers),
		}
		if cfg.Otel.Insecure {
			opts = append(opts, otlploggrpc.WithDialOption(
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			))
		}
		exp, err = otlploggrpc.New(context.Background(), opts...)

	case "http":
		opts := []otlploghttp.Option{
			otlploghttp.WithEndpoint(cfg.Otel.Endpoint),
			otlploghttp.WithURLPath(fmt.Sprintf("/api/%s/v1/logs", cfg.Otel.Organization)),
			otlploghttp.WithHeaders(headers),
		}
		if cfg.Otel.Insecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		exp, err = otlploghttp.New(context.Background(), opts...)

	default:
		return nil, fmt.Errorf("unsupported otel.protocol: %s", cfg.Otel.Protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	return exp, nil
}

// CreateResource 创建 otel resource
func CreateResource(cfg Config) (*resource.Resource, error) {
	return resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
}

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
