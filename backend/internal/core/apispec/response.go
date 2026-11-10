package apispec

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	HeaderRequestID = "x-request-id"
	HeaderTraceID   = "x-trace-id"
	HeaderLatencyMS = "x-latency-ms"

	ctxStartTime = "apispec.start"
	ctxRequestID = "apispec.request_id"
)

// ResponseBase 是所有 HTTP/JSON 接口的统一响应。
type ResponseBase struct {
	Code     int32        `json:"code"`
	CodeName string       `json:"code_name"`
	Message  string       `json:"message"`
	Data     any          `json:"data"`
	Meta     Meta         `json:"meta"`
	Error    *ErrorDetail `json:"error,omitempty"`
}

type Meta struct {
	RequestID string `json:"request_id"`
	LatencyMS int32  `json:"latency_ms"`
	TraceID   string `json:"trace_id,omitempty"`
}

type ErrorDetail struct {
	Reason            string            `json:"reason"`
	Retryable         bool              `json:"retryable"`
	RetryAfterSeconds int32             `json:"retry_after_seconds,omitempty"`
	FieldViolations   []FieldViolation  `json:"field_violations,omitempty"`
	DocURL            string            `json:"doc_url,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

var inboundRequestID = regexp.MustCompile(`^[A-Za-z0-9._:-]{8,128}$`)

// Observe 记录请求开始时间并确定 request id；须挂在契约接口的路由组最前面。
func Observe() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ctxStartTime, time.Now())
		id := c.GetHeader(HeaderRequestID)
		if !inboundRequestID.MatchString(id) {
			id = NewSID("req")
		}
		c.Set(ctxRequestID, id)
		c.Next()
	}
}

// RequestID 返回当前请求标识。
func RequestID(c *gin.Context) string { return c.GetString(ctxRequestID) }

// WriteOK 写出成功响应；data 为 nil 时输出空对象。
func WriteOK(c *gin.Context, data any) {
	if data == nil {
		data = struct{}{}
	}
	write(c, http.StatusOK, ResponseBase{Code: 0, CodeName: "", Message: "success", Data: data})
}

// WriteError 是唯一的错误出口：domain error 原样映射，其余错误按内部错误处理并记录日志。
func WriteError(c *gin.Context, err error) {
	e := AsError(err)
	if e.Code.Status >= http.StatusInternalServerError {
		zap.L().Error("api request failed",
			zap.String("request_id", RequestID(c)), zap.String("path", c.FullPath()),
			zap.String("reason", e.Reason), zap.Error(err))
	}
	if e.Code == CodeRateLimited || e.Code == CodeQuotaExceeded {
		if e.RetryAfterSeconds > 0 {
			c.Header("Retry-After", strconv.Itoa(int(e.RetryAfterSeconds)))
		}
	}
	violations := e.FieldViolations
	write(c, e.Code.Status, ResponseBase{
		Code: e.Code.Value, CodeName: e.Code.Name, Message: e.Message, Data: nil,
		Error: &ErrorDetail{
			Reason: e.Reason, Retryable: e.Retryable, RetryAfterSeconds: e.RetryAfterSeconds,
			FieldViolations: violations, Metadata: e.Metadata,
		},
	})
}

// AbortWithError 用于中间件：写出错误并终止后续处理。
func AbortWithError(c *gin.Context, err error) {
	WriteError(c, err)
	c.Abort()
}

func write(c *gin.Context, status int, body ResponseBase) {
	body.Meta = meta(c)
	c.Header(HeaderRequestID, body.Meta.RequestID)
	c.Header(HeaderLatencyMS, strconv.Itoa(int(body.Meta.LatencyMS)))
	if body.Meta.TraceID != "" {
		c.Header(HeaderTraceID, body.Meta.TraceID)
	}
	c.JSON(status, body)
}

func meta(c *gin.Context) Meta {
	m := Meta{RequestID: RequestID(c)}
	if m.RequestID == "" {
		m.RequestID = NewSID("req")
		c.Set(ctxRequestID, m.RequestID)
	}
	if start, ok := c.Get(ctxStartTime); ok {
		m.LatencyMS = int32(time.Since(start.(time.Time)).Milliseconds())
	}
	if sc := trace.SpanContextFromContext(c.Request.Context()); sc.HasTraceID() {
		m.TraceID = sc.TraceID().String()
	}
	return m
}
