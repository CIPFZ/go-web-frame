package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// 限制最大记录长度 (1MB)，防止内存爆炸，即便 DB 能存，内存队列也不建议存太大
const maxBodyLogSize = 1024 * 1024

// ✨ 对象池：复用 Response Body 的 Buffer，减少 GC
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096)) // 默认 4KB
	},
}

func OperationRecord(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 过滤非修改类请求
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// 2. 读取 Request Body
		var reqBody []byte
		if c.Request.Body != nil {
			// 这里的 ReadAll 是必须的，因为 binding 需要读
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody)) // 回写
		}

		// 3. 包装 ResponseWriter 以捕获响应
		// 从池中获取 Buffer
		respBuf := bufferPool.Get().(*bytes.Buffer)
		respBuf.Reset()
		defer bufferPool.Put(respBuf) // 归还 Buffer

		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           respBuf,
		}
		c.Writer = writer

		// 4. 记录开始时间
		start := time.Now()

		// --- 执行业务 ---
		c.Next()
		// -------------

		// 5. 异步处理日志 (避免阻塞 API 响应)
		// 注意：这里需要拷贝一份数据，因为 c.Request 在请求结束后会被 Gin 重置
		// 为了性能，我们只拷贝需要的数据

		// 提取 Trace 信息
		var traceId, spanId string
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			traceId = span.SpanContext().TraceID().String()
			spanId = span.SpanContext().SpanID().String()
		}

		// 截断 Body (保护内存)
		reqBodyStr := truncateString(string(reqBody), maxBodyLogSize)
		respBodyStr := truncateString(writer.body.String(), maxBodyLogSize)

		// 解码 Path (处理中文路径)
		path, _ := url.QueryUnescape(c.Request.URL.Path)

		record := model.SysOperationLog{
			Ip:       c.ClientIP(),
			Method:   c.Request.Method,
			Path:     path,
			Agent:    c.Request.UserAgent(),
			Body:     reqBodyStr,
			Resp:     respBodyStr,
			Status:   c.Writer.Status(),
			Latency:  time.Since(start),
			UserID:   utils.GetUserID(c), // 假设 utils 已经打磨好
			TraceID:  traceId,
			SpanID:   spanId,
			ErrorMsg: c.Errors.String(), // 可选
		}

		// 推入队列
		if svcCtx.AuditRecorder != nil {
			svcCtx.AuditRecorder.Push(record)
		}
	}
}

// truncateString 字符串截断辅助函数
func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max] + "...(truncated)"
	}
	return s
}

// responseBodyWriter 包装器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// WriteString 必须重写，因为 Gin 有时调用 WriteString
func (r *responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}
