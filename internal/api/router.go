package api

import (
	"fmt"
	"github.com/CIPFZ/gowebframe/internal/api/response"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"time"
)

// Package api -----------------------------
// @file        : router.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:50
// @description :
// -------------------------------------------

func RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		response.OkWithMessage("pong", c)
	})

	r.GET("/metrics-prom", gin.WrapH(promhttp.Handler()))

	r.GET("/hello", func(c *gin.Context) {
		time.Sleep(150 * time.Millisecond) // 模拟耗时
		response.OkWithMessage("Hello, OpenTelemetry!", c)
	})

	r.GET("/log", func(c *gin.Context) {
		// 1. 从 context 获取由中间件创建的 logger
		log := logger.GetLogger(c)
		log.Info("This is a test struct info log", zap.String("id", "ID:123"), zap.Int("id", 123), zap.Bool("bool", true))
		log.Error("This is a error test log", zap.Error(fmt.Errorf("this is an error")))
		response.OkWithMessage("Hello, Log!", c)
	})
}
