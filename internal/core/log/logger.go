package logger

import (
	"context"
	"fmt"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Package logger -----------------------------
// @file        : logger.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:24
// @description :
// -------------------------------------------

// NewLogger 根据配置创建 zap.Logger
func NewLogger(cfg *config.Logger, otelCfg *config.OTELLoggerConfig) (*zap.Logger, error) {
	// 解析日志级别
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(cfg.Level)); err != nil {
		lvl = zapcore.InfoLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
	}

	// 支持 console | json
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	// outputs
	var cores []zapcore.Core

	// helper to create writer for file using lumberjack
	if cfg.Output == "file" || cfg.Output == "both" {
		if cfg.FilePath == "" {
			cfg.FilePath = "logs/app.log"
		}
		_ = os.MkdirAll(filepath.Dir(cfg.FilePath), 0755)

		if cfg.MaxSizeMB == 0 {
			cfg.MaxSizeMB = 100
		}
		if cfg.MaxBackups == 0 {
			cfg.MaxBackups = 7
		}
		if cfg.MaxAgeDays == 0 {
			cfg.MaxAgeDays = 30
		}
		lumber := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSizeMB, // megabytes
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays, // days
			Compress:   cfg.Compress,
		}
		writer := zapcore.AddSync(lumber)
		core := zapcore.NewCore(encoder, writer, lvl)
		cores = append(cores, core)
	}

	// stdout
	if cfg.Output == "stdout" || cfg.Output == "both" {
		stdoutWriter := zapcore.Lock(os.Stdout)
		core := zapcore.NewCore(encoder, stdoutWriter, lvl)
		cores = append(cores, core)
	}

	// fallback if no core configured
	if len(cores) == 0 {
		// default to stdout
		stdoutWriter := zapcore.Lock(os.Stdout)
		core := zapcore.NewCore(encoder, stdoutWriter, lvl)
		cores = append(cores, core)
	}

	// 如果 ooProvider 不为 nil，就追加一个 core
	if otelCfg != nil && otelCfg.LogProvider != nil {
		ooCore := NewOTELLogCore(otelCfg, lvl)
		cores = append(cores, ooCore)
	}

	core := zapcore.NewTee(cores...)

	// optional sampling
	if cfg.EnableSample {
		core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
	}

	return zap.New(core, zap.AddCaller()), nil
}

// Sync flushes any buffered logs (调用时请忽略返回错误)
func Sync(l *zap.Logger) {
	_ = l.Sync()
}

// GetLogger 支持从 gin.Context 或 context.Context 中获取
func GetLogger(ctx interface{}) *zap.Logger {
	var logger interface{}

	switch c := ctx.(type) {
	case *gin.Context:
		// 优先从 Gin Keys 取
		if v, exists := c.Get(utils.LoggerKey); exists {
			logger = v
		} else {
			// 兜底：如果 Gin Keys 没有，尝试从 Request.Context 取
			logger = c.Request.Context().Value(utils.LoggerKey)
		}
	case context.Context:
		// 从标准 Context 取
		logger = c.Value(utils.LoggerKey)
	}

	// 只要 logger 不为 nil 且类型断言成功，就返回
	if logger != nil {
		if l, ok := logger.(*zap.Logger); ok {
			return l
		}
	}
	return zap.L()
}

// Helper: Create an io.Writer that writes to both stdout and an io.Writer (unused now, kept for extension)
func multiWriter(writers ...io.Writer) io.Writer {
	if len(writers) == 0 {
		return os.Stdout
	}
	if len(writers) == 1 {
		return writers[0]
	}
	return io.MultiWriter(writers...)
}

func debugStack() {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		fmt.Printf("stack frame: %s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
}

// ooCore 是一个 zapcore.Core 封装，负责把 zap 日志转发到 OpenObserve (OTLP logs)
type otelLogCore struct {
	levelEnabler zapcore.LevelEnabler
	lp           *sdklog.LoggerProvider
	fields       []zapcore.Field
	otelLogger   log.Logger
	serviceName  string
	serviceVer   string
	environment  string
}

func NewOTELLogCore(otelCfg *config.OTELLoggerConfig, lvl zapcore.LevelEnabler) zapcore.Core {
	return &otelLogCore{
		levelEnabler: lvl,
		lp:           otelCfg.LogProvider,
		fields:       make([]zapcore.Field, 0),
		otelLogger:   otelCfg.LogProvider.Logger("zap", log.WithInstrumentationVersion("1.0.0")),
		serviceName:  otelCfg.ServiceName,
		serviceVer:   otelCfg.ServiceVer,
		environment:  otelCfg.Environment,
	}
}

func (c *otelLogCore) Enabled(lvl zapcore.Level) bool {
	return c.levelEnabler.Enabled(lvl)
}

func (c *otelLogCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.fields = append(make([]zapcore.Field, 0, len(c.fields)+len(fields)), c.fields...)
	clone.fields = append(clone.fields, fields...)
	return &clone
}

func (c *otelLogCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

// Write 将 zap 日志条目转换为结构化的 OTel 日志记录
func (c *otelLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	ctx := context.Background()

	// 创建一个新的 OTLP 记录
	rec := log.Record{}
	rec.SetTimestamp(entry.Time)
	rec.SetObservedTimestamp(time.Now())
	rec.SetSeverity(zapLevelToSeverity(entry.Level))
	rec.SetSeverityText(entry.Level.String())

	// body 只放 entry.Message
	rec.SetBody(log.StringValue(entry.Message))

	// 合并 With 字段和当前字段
	allFields := append(c.fields, fields...)
	attrs := make([]log.KeyValue, 0, len(allFields)+6)

	// resource 固定字段
	attrs = append(attrs,
		log.String("service.name", c.serviceName),
		log.String("service.version", c.serviceVer),
		log.String("deployment.environment", c.environment),
	)

	// caller 信息
	if entry.Caller.Defined {
		attrs = append(attrs,
			log.String("code.filepath", entry.Caller.File),
			log.Int("code.lineno", entry.Caller.Line),
		)
		// 函数名可能为空（zap 默认不一定填），需要容错
		if entry.Caller.Function != "" {
			attrs = append(attrs, log.String("code.function", entry.Caller.Function))
		}
	}

	// zap field 转 attributes
	for _, field := range allFields {
		attrs = append(attrs, zapFieldToOtelKV(field))
	}

	if len(attrs) > 0 {
		rec.AddAttributes(attrs...)
	}

	// 上报
	c.otelLogger.Emit(ctx, rec)
	return nil
}

func (c *otelLogCore) Sync() error {
	return nil
}

// zapFieldToOtelKV 是一个辅助函数，将 zapcore.Field 转换为 log.KeyValue
func zapFieldToOtelKV(field zapcore.Field) log.KeyValue {
	switch field.Type {
	case zapcore.BoolType:
		return log.Bool(field.Key, field.Integer == 1)
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type, zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
		return log.Int64(field.Key, field.Integer)
	case zapcore.Float64Type, zapcore.Float32Type:
		return log.Float64(field.Key, field.Interface.(float64))
	case zapcore.StringType:
		return log.String(field.Key, field.String)
	case zapcore.StringerType:
		return log.String(field.Key, field.Interface.(fmt.Stringer).String())
	case zapcore.DurationType:
		// OTel 推荐将 duration 存为纳秒的 int64
		return log.Int64(field.Key, field.Integer)
	case zapcore.ErrorType:
		return log.String(field.Key, field.Interface.(error).Error())
	// 其他类型可以根据需要添加，例如 ArrayType, ObjectType 等
	default:
		// 对于不支持的复杂类型，使用 fmt.Sprintf 作为回退
		return log.String(field.Key, fmt.Sprintf("%+v", field.Interface))
	}
}

// 映射 zap 的 Level → OTel Severity
func zapLevelToSeverity(l zapcore.Level) log.Severity {
	switch l {
	case zapcore.DebugLevel:
		return log.SeverityDebug
	case zapcore.InfoLevel:
		return log.SeverityInfo
	case zapcore.WarnLevel:
		return log.SeverityWarn
	case zapcore.ErrorLevel:
		return log.SeverityError
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return log.SeverityFatal
	default:
		return log.SeverityUndefined
	}
}
