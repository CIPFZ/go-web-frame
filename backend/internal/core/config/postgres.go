package config

import (
	"fmt"
	"strings"

	"gorm.io/gorm/logger"
)

type Postgres struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	Dbname       string `mapstructure:"db_name" json:"db_name" yaml:"db_name"`
	Username     string `mapstructure:"username" json:"username" yaml:"username"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	SSLMode      string `mapstructure:"ssl_mode" json:"ssl_mode" yaml:"ssl_mode"`
	TimeZone     string `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
	LogMode      string `mapstructure:"log_mode" json:"log_mode" yaml:"log_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	Singular     bool   `mapstructure:"singular" json:"singular" yaml:"singular"`
	LogZap       bool   `mapstructure:"log_zap" json:"log_zap" yaml:"log_zap"`
}

func (p Postgres) LogLevel() logger.LogLevel {
	switch strings.ToLower(p.LogMode) {
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

func (p Postgres) Dsn() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		p.Host,
		p.Username,
		p.Password,
		p.Dbname,
		p.Port,
		p.SSLMode,
		p.TimeZone,
	)
}
