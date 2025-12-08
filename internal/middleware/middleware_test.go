package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// performRequest 辅助函数
func performRequest(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestRateLimitMiddleware 测试限流中间件
func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// QPS=1, Burst=1
	cfg := config.RateLimitConfig{Enabled: true, QPS: 1, Burst: 1}

	r := gin.New()
	r.Use(RateLimitMiddleware(cfg))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 1. 成功
	w1 := performRequest(r, "GET", "/ping")
	assert.Equal(t, http.StatusOK, w1.Code)

	// 2. 失败 (429)
	w2 := performRequest(r, "GET", "/ping")
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

// TestBreakerMiddleware 测试熔断中间件
func TestBreakerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	r := gin.New()

	// 注入熔断中间件
	// 添加 Debug 日志以便观察
	r.Use(func(c *gin.Context) {
		fmt.Println("👉 [Test] 进入中间件")
		c.Next()
		fmt.Printf("👈 [Test] 退出中间件, Status: %d\n", c.Writer.Status())
	})
	r.Use(BreakerMiddleware(logger))

	// 1. 正常路由
	r.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "ok"})
	})

	// 2. 必挂路由
	r.GET("/fail", func(c *gin.Context) {
		fmt.Println("💥 [Test] 执行 /fail Handler")

		// ✨✨✨ 终极修复：强制立即写入 Header ✨✨✨
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteHeaderNow()
		// 或者 c.String(500, "error") 也会触发
	})

	fmt.Println("--- Request A: Success ---")
	w1 := performRequest(r, "GET", "/success")
	assert.Equal(t, http.StatusOK, w1.Code)

	fmt.Println("--- Request B: Loop Fail ---")
	// 触发熔断
	for i := 0; i < 15; i++ { //稍微增加次数确保触发
		wFail := performRequest(r, "GET", "/fail")

		// ✨✨✨ 智能断言 ✨✨✨
		// 这里的请求有两种可能：
		// 1. 熔断器没开：执行 Handler -> 返回 500
		// 2. 熔断器开了：中间件拦截 -> 返回 503 (修复后)
		if wFail.Code != http.StatusInternalServerError && wFail.Code != http.StatusServiceUnavailable {
			t.Errorf("Request %d: expected 500 or 503, got %d", i, wFail.Code)
		}
	}

	fmt.Println("--- Request C: Verify Open ---")
	// 验证熔断开启
	wBlocked := performRequest(r, "GET", "/success")

	// 这里必须是 503
	assert.Equal(t, http.StatusServiceUnavailable, wBlocked.Code, "Breaker should be open")
	assert.Contains(t, wBlocked.Body.String(), "Breaker Open")
}
