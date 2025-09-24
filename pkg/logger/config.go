package logger

import sdklog "go.opentelemetry.io/otel/sdk/log"

// Package logger -----------------------------
// @file        : config.go
// @author      : CIPFZ
// @time        : 2025/9/20 15:53
// @description :
// -------------------------------------------

// Config 是 NewLogger 的配置结构体
type Config struct {
	Level        string `json:"level" yaml:"level" toml:"level" mapstructure:"level"`                         // "debug","info","warn","error"
	Output       string `json:"output" yaml:"output" toml:"output" mapstructure:"output"`                     // "stdout" | "file" | "both"
	Format       string `json:"format" yaml:"format" toml:"format" mapstructure:"format"`                     // "json" | "console"
	FilePath     string `json:"file_path" yaml:"file_path" toml:"file_path" mapstructure:"file_path"`         // 如果使用 file 或 both
	MaxSizeMB    int    `json:"max_size_mb" yaml:"max_size_mb" toml:"max_size_mb" mapstructure:"max_size_mb"` // lumberjack
	MaxBackups   int    `json:"max_backups" yaml:"max_backups" toml:"max_backups" mapstructure:"max_backups"`
	MaxAgeDays   int    `json:"max_age_days" yaml:"max_age_days" toml:"max_age_days" mapstructure:"max_age_days"`
	Compress     bool   `json:"compress" yaml:"compress" toml:"compress" mapstructure:"compress"`
	EnableCaller bool   `json:"enable_caller" yaml:"enable_caller" toml:"enable_caller" mapstructure:"enable_caller"`
	EnableSample bool   `json:"enable_sample" yaml:"enable_sample" toml:"enable_sample" mapstructure:"enable_sample"` // 是否启用采样，减低高频日志压力
}

// OTELLoggerConfig 日志通过 OTEL 进行发送所需要的配置信息
type OTELLoggerConfig struct {
	LogProvider *sdklog.LoggerProvider
	ServiceName string
	ServiceVer  string
	Environment string
}
