package i18n

// Package i18n -----------------------------
// @file        : config.go
// @author      : CIPFZ
// @time        : 2025/9/20 15:58
// @description :
// -------------------------------------------

// Config 用于 i18n 初始化的配置
type Config struct {
	Path string `json:"path" yaml:"path" toml:"path" mapstructure:"path"` // 翻译文件目录
}
