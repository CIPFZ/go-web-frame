package middleware

import (
	"net/http"
	"sync"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter IP限流器管理器
type IPRateLimiter struct {
	ips   sync.Map
	qps   rate.Limit
	burst int
}

func NewIPRateLimiter(qps int, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		qps:   rate.Limit(qps),
		burst: burst,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	limiter, exists := i.ips.Load(ip)
	if !exists {
		// 创建一个新的限流器
		newLimiter := rate.NewLimiter(i.qps, i.burst)
		// 使用 LoadOrStore 避免并发创建覆盖
		actual, _ := i.ips.LoadOrStore(ip, newLimiter)
		return actual.(*rate.Limiter)
	}
	return limiter.(*rate.Limiter)
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	limiterManager := NewIPRateLimiter(cfg.QPS, cfg.Burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiterManager.GetLimiter(ip)

		// Allow() 是非阻塞的，如果桶里没令牌，返回 false
		if !limiter.Allow() {
			// 429 Too Many Requests
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 7, // 或者其他你约定的错误码
				"msg":  "访问过于频繁，请稍后再试",
				"data": nil,
			})
			return
		}

		c.Next()
	}
}
