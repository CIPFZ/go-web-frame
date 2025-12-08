package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

// ZapGormLogger 适配器：将 GORM 日志桥接到 Zap
type ZapGormLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

// NewZapGormLogger 构造函数
func NewZapGormLogger(zapLogger *zap.Logger, cfg config.MySQL) *ZapGormLogger {
	// 映射配置中的 LogLevel 到 GORM LogLevel
	var lvl gormlogger.LogLevel
	switch cfg.LogMode {
	case "silent":
		lvl = gormlogger.Silent
	case "error":
		lvl = gormlogger.Error
	case "warn":
		lvl = gormlogger.Warn
	case "info":
		lvl = gormlogger.Info
	default:
		lvl = gormlogger.Info
	}

	return &ZapGormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  lvl,
		SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
		IgnoreRecordNotFoundError: true,                   // 忽略 RecordNotFound 错误 (通常视为业务逻辑而非系统错误)
	}
}

// LogMode 实现 Interface：设置日志级别
func (l *ZapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 实现 Interface
func (l *ZapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	fmt.Printf("get logger ctx: %+v", logger.GetLogger(ctx))
	if l.LogLevel >= gormlogger.Info {
		// ✨ 关键：使用 logger.GetLogger(ctx) 获取带有 TraceID 的 Logger
		logger.GetLogger(ctx).Info(fmt.Sprintf(msg, data...), zap.String("module", "gorm"))
	}
}

// Warn 实现 Interface
func (l *ZapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		logger.GetLogger(ctx).Warn(fmt.Sprintf(msg, data...), zap.String("module", "gorm"))
	}
}

// Error 实现 Interface
func (l *ZapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		logger.GetLogger(ctx).Error(fmt.Sprintf(msg, data...), zap.String("module", "gorm"))
	}
}

// Trace 实现 Interface：核心 SQL 记录逻辑
func (l *ZapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)

	// 获取带 TraceID 的 Logger
	log := logger.GetLogger(ctx).With(
		zap.String("module", "gorm"),
		zap.Duration("duration", elapsed),
	)

	// 获取 SQL 和行数
	sql, rows := fc()
	if rows != -1 {
		log = log.With(zap.Int64("rows", rows))
	}

	// 1. 处理错误
	if err != nil && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)) {
		log.Error("sql_error", zap.String("sql", sql), zap.Error(err))
		return
	}

	// 2. 处理慢查询
	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		log.Warn("sql_slow", zap.String("sql", sql))
		return
	}

	// 3. 记录普通 SQL (仅当级别为 Info 时)
	if l.LogLevel == gormlogger.Info {
		log.Info("sql_exec", zap.String("sql", sql))
	}
}
