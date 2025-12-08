package observability

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// HTTPMetrics 结构体包含了 Otel 规范中 HTTP Server 的四个核心指标
type HTTPMetrics struct {
	activeRequests   metric.Int64UpDownCounter
	requestDuration  metric.Float64Histogram
	requestBodySize  metric.Int64Histogram // <-- 新增
	responseBodySize metric.Int64Histogram // <-- 新增
}

// NewHTTPMetrics 使用 Otel Meter 创建和注册 HTTP 指标
func NewHTTPMetrics(meter metric.Meter) (*HTTPMetrics, error) {
	var err error
	m := &HTTPMetrics{}

	// 1. 瞬时并发请求数 (UpDownCounter)
	m.activeRequests, err = meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// 2. 请求耗时 (Histogram)
	m.requestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("s"), // "s" = 秒
	)
	if err != nil {
		return nil, err
	}

	// 3. ✨ 新增：请求体大小 (Histogram)
	m.requestBodySize, err = meter.Int64Histogram(
		"http.server.request.body.size",
		metric.WithDescription("Size of HTTP request bodies"),
		metric.WithUnit("By"), // "By" = 字节
	)
	if err != nil {
		return nil, err
	}

	// 4. ✨ 新增：响应体大小 (Histogram)
	m.responseBodySize, err = meter.Int64Histogram(
		"http.server.response.body.size",
		metric.WithDescription("Size of HTTP response bodies"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// 5. 移除了多余的 "http.server.request.count"

	return m, nil
}

// Middleware 返回一个 Gin 中间件，用于自动记录指标
func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		start := time.Now()

		// --- 1. 获取低基数路由 ---
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		// --- 2. 记录请求开始时的指标 (Active Requests, Request Size) ---
		// (这些属性在请求开始时就知道)
		preAttrs := attribute.NewSet(
			semconv.HTTPRequestMethodKey.String(c.Request.Method),
			semconv.HTTPRouteKey.String(route),
		)

		m.activeRequests.Add(ctx, 1, metric.WithAttributeSet(preAttrs))
		m.requestBodySize.Record(ctx, c.Request.ContentLength, metric.WithAttributeSet(preAttrs))

		// --- 3. 延迟“瞬时并发请求” (在处理后 -1) ---
		defer m.activeRequests.Add(ctx, -1, metric.WithAttributeSet(preAttrs))

		// --- 4. 执行处理器 (c.Next) ---
		c.Next()

		// --- 5. 记录请求结束时的指标 (Duration, Response Size) ---
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		responseSize := c.Writer.Size()
		if responseSize < 0 {
			responseSize = 0
		}

		// (这些属性在请求结束后才知道)
		finalAttrs := attribute.NewSet(
			semconv.HTTPRequestMethodKey.String(c.Request.Method),
			semconv.HTTPRouteKey.String(route),
			semconv.HTTPResponseStatusCodeKey.Int(status),
		)

		m.requestDuration.Record(ctx, duration, metric.WithAttributeSet(finalAttrs))
		m.responseBodySize.Record(ctx, int64(responseSize), metric.WithAttributeSet(finalAttrs))
	}
}
