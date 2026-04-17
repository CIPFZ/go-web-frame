package config

import (
	"path/filepath"
	"strings"

	"gorm.io/gorm/logger"
)

type SQLite struct {
	Path          string `mapstructure:"path" json:"path" yaml:"path"`
	WAL           bool   `mapstructure:"wal" json:"wal" yaml:"wal"`
	BusyTimeoutMS int    `mapstructure:"busy_timeout_ms" json:"busy_timeout_ms" yaml:"busy_timeout_ms"`
	ForeignKeys   bool   `mapstructure:"foreign_keys" json:"foreign_keys" yaml:"foreign_keys"`
	LogMode       string `mapstructure:"log_mode" json:"log_mode" yaml:"log_mode"`
	MaxIdleConns  int    `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns  int    `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	Singular      bool   `mapstructure:"singular" json:"singular" yaml:"singular"`
	LogZap        bool   `mapstructure:"log_zap" json:"log_zap" yaml:"log_zap"`
}

func (s SQLite) LogLevel() logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(s.LogMode)) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

func (s SQLite) DBName() string {
	if strings.TrimSpace(s.Path) == "" {
		return "sqlite3"
	}
	return filepath.Base(strings.TrimSpace(s.Path))
}
