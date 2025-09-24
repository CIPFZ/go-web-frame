package logger

// Package logger -----------------------------
// @file        : logger_test.go
// @author      : CIPFZ
// @time        : 2025/9/20 14:54
// @description :
// -------------------------------------------

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// newBufferLogger 创建一个写到 bytes.Buffer 的 zap.Logger，用于断言输出
func newBufferLogger() (*zap.Logger, *bytes.Buffer) {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.DebugLevel)
	l := zap.New(core)
	return l, buf
}

func TestWithTrace_IncludesTraceIDs(t *testing.T) {
	base, buf := newBufferLogger()

	// 使用 SDK tracer 生成有效 span context
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "span")
	defer span.End()

	l := WithTrace(ctx, base)
	l.Info("hello trace test")
	out := buf.String()

	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		t.Fatalf("expected valid span context")
	}

	if !strings.Contains(out, sc.TraceID().String()) {
		t.Fatalf("trace_id not found in log output: %s", out)
	}
	if !strings.Contains(out, sc.SpanID().String()) {
		t.Fatalf("span_id not found in log output: %s", out)
	}
}

func TestGinLoggerMiddleware_SetsLoggerAndLogs(t *testing.T) {
	base, buf := newBufferLogger()

	// create valid trace context
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "span")
	defer span.End()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(ctx)
	c.Request = req

	mw := GinLoggerMiddleware(base)
	mw(c)

	// 检查 gin context 中是否注入 logger
	v, ok := c.Get("logger")
	if !ok {
		t.Fatalf("logger not found in gin context")
	}
	if _, ok := v.(*zap.Logger); !ok {
		t.Fatalf("logger in context is of wrong type")
	}

	// 检查 middleware 的访问日志写入并包含 trace_id
	out := buf.String()
	if !strings.Contains(out, "http_request") {
		t.Fatalf("expected http_request in logs, got: %s", out)
	}

	sc := trace.SpanFromContext(ctx).SpanContext()
	if !strings.Contains(out, sc.TraceID().String()) {
		t.Fatalf("trace id not found in middleware log: %s", out)
	}
}
