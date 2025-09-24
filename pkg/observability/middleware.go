package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Package observability -----------------------------
// @file        : middleware.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:49
// @description :
// -------------------------------------------

type HTTPMetrics struct {
	requestCount    metric.Int64Counter
	requestDuration metric.Float64Histogram
}

func NewHTTPMetrics(m metric.Meter) (*HTTPMetrics, error) {
	reqCount, err := m.Int64Counter("http.server.request_count",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	reqDuration, err := m.Float64Histogram("http.server.request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
	)
	if err != nil {
		return nil, err
	}

	return &HTTPMetrics{
		requestCount:    reqCount,
		requestDuration: reqDuration,
	}, nil
}

func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		status := strconv.Itoa(c.Writer.Status())

		attrs := []attribute.KeyValue{
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route),
			attribute.String("http.status_code", status),
		}

		m.requestCount.Add(c.Request.Context(), 1, metric.WithAttributes(attrs...))
		m.requestDuration.Record(c.Request.Context(), duration, metric.WithAttributes(attrs...))
	}
}
