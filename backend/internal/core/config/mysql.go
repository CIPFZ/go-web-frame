package config

import (
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm/logger"
)

type MySQL struct {
	Prefix       string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	Config       string `mapstructure:"config" json:"config" yaml:"config"`
	Dbname       string `mapstructure:"db_name" json:"db_name" yaml:"db_name"`
	Username     string `mapstructure:"username" json:"username" yaml:"username"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Path         string `mapstructure:"path" json:"path" yaml:"path"`
	Engine       string `mapstructure:"engine" json:"engine" yaml:"engine" default:"InnoDB"`
	LogMode      string `mapstructure:"log_mode" json:"log_mode" yaml:"log_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	Singular     bool   `mapstructure:"singular" json:"singular" yaml:"singular"`
	LogZap       bool   `mapstructure:"log_zap" json:"log_zap" yaml:"log_zap"`
}

func (m MySQL) LogLevel() logger.LogLevel {
	switch strings.ToLower(m.LogMode) {
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

func (m MySQL) Dsn() string {
	host := strings.TrimSpace(m.Host)
	if host == "" {
		host = strings.TrimSpace(m.Path)
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		m.Username,
		m.Password,
		host,
		m.Port,
		m.Dbname,
		normalizeMySQLQuery(m.Config),
	)
}

func normalizeMySQLQuery(raw string) string {
	defaults := map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "True",
		"loc":       "Local",
	}

	if strings.TrimSpace(raw) == "" {
		v := url.Values{}
		for key, value := range defaults {
			v.Set(key, value)
		}
		return v.Encode()
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		normalized := raw
		for key, value := range defaults {
			if !strings.Contains(strings.ToLower(normalized), strings.ToLower(key)+"=") {
				if !strings.HasSuffix(normalized, "&") {
					normalized += "&"
				}
				normalized += key + "=" + value
			}
		}
		return normalized
	}

	for key, value := range defaults {
		if strings.TrimSpace(values.Get(key)) == "" {
			values.Set(key, value)
		}
	}
	return values.Encode()
}
