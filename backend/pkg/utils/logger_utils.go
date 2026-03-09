package utils

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const LoggerKey = "zap_logger"

// GetLogger 辅助函数也移到这里 (或者放在 pkg/utils/logger_utils.go)
func GetLogger(ctx interface{}) *zap.Logger {
	var logger interface{}
	var ok bool

	switch c := ctx.(type) {
	case *gin.Context:
		logger, ok = c.Get(LoggerKey)
	case context.Context:
		logger = c.Value(LoggerKey)
	default:
		return zap.L()
	}

	if ok && logger != nil {
		if l, ok := logger.(*zap.Logger); ok {
			return l
		}
	}
	return zap.L()
}
