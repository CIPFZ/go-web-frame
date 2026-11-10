package middleware

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/audit"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// 业务模块可在记录前写入这两个键，补充操作日志的模块与说明。
const (
	CtxKeyAuditModule = "audit.module"
	CtxKeyAuditRemark = "audit.remark"
	// CtxKeyAuditOmitBodies 为真时只记录“谁调用了什么”，不记录请求与响应正文（如返回秘密值的接口）
	CtxKeyAuditOmitBodies = "audit.omit_bodies"
)

// Capture at most one extra byte so oversized bodies can be omitted altogether.
// Returning len(p) lets TeeReader stream the original body without changing it.
type boundedCapture struct{ bytes.Buffer }

func (b *boundedCapture) Write(p []byte) (int, error) {
	n := len(p)
	remaining := audit.MaxBodySize + 1 - b.Len()
	if remaining > 0 {
		_, _ = b.Buffer.Write(p[:min(remaining, len(p))])
	}
	return n, nil
}

type capturedRequest struct {
	io.Reader
	io.Closer
}

func OperationRecord(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svcCtx.AuditRecorder == nil || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		var requestBody boundedCapture
		mediaType, _, _ := mime.ParseMediaType(c.GetHeader("Content-Type"))
		isJSON := mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
		if c.Request.Body != nil && isJSON {
			original := c.Request.Body
			c.Request.Body = &capturedRequest{Reader: io.TeeReader(original, &requestBody), Closer: original}
		}
		writer := &responseBodyWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		start := time.Now()
		c.Next()

		var traceID, spanID string
		sc := trace.SpanContextFromContext(c.Request.Context())
		if sc.IsValid() {
			traceID, spanID = sc.TraceID().String(), sc.SpanID().String()
		}
		errorSummary := ""
		if len(c.Errors) > 0 {
			errorSummary = fmt.Sprintf("%d handler error(s)", len(c.Errors))
		}
		svcCtx.AuditRecorder.Push(model.SysOperationLog{
			Ip: c.ClientIP(), Method: c.Request.Method,
			Path:  truncateString(c.Request.URL.Path, 2048),
			Agent: truncateString(c.Request.UserAgent(), 512),
			Body:  bodyOf(c, requestBody.Bytes()), Resp: bodyOf(c, writer.body.Bytes()),
			Status: c.Writer.Status(), Latency: time.Since(start),
			UserID: utils.GetUserID(c), TokenID: tokenOf(c), TraceID: traceID, SpanID: spanID, ErrorMsg: errorSummary,
			Module: truncateString(c.GetString(CtxKeyAuditModule), 64), Remark: truncateString(c.GetString(CtxKeyAuditRemark), 128),
		})
	}
}

// bodyOf 是记录进操作日志的正文：按字段名脱敏；接口声明不记录正文时只留占位。
func bodyOf(c *gin.Context, body []byte) string {
	if c.GetBool(CtxKeyAuditOmitBodies) {
		return "[body omitted: contains secrets]"
	}
	return audit.SanitizeBody(body)
}

// tokenOf 是这次操作经过的 API Token（网页操作为空）。
func tokenOf(c *gin.Context) *uint {
	if id := c.GetUint(CtxKeyAPITokenID); id != 0 {
		return &id
	}
	return nil
}

func truncateString(s string, max int) string {
	if len(s) > max {
		return strings.ToValidUTF8(s[:max], "")
	}
	return s
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body boundedCapture
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	_, _ = r.body.Write(b[:n])
	return n, err
}

func (r *responseBodyWriter) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}
