package config

import (
	"fmt"
	"github.com/CIPFZ/gowebframe/pkg/i18n"
	"github.com/CIPFZ/gowebframe/pkg/observability"

	"github.com/CIPFZ/gowebframe/pkg/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Package config -----------------------------
// @file        : config.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:24
// @description :
// -------------------------------------------

// AppConfig 应用配置
type AppConfig struct {
	Name        string `yaml:"name" mapstructure:"name"`
	Mode        string `yaml:"mode" mapstructure:"mode"`
	Environment string `yaml:"environment" mapstructure:"environment"`
}

// ServerConfig 服务相关配置
type ServerConfig struct {
	Port int    `yaml:"port" mapstructure:"port"`
	Mode string `yaml:"mode" mapstructure:"mode"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Enable bool   `yaml:"enable" mapstructure:"enable"`
	DSN    string `yaml:"dsn" mapstructure:"dsn"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Enable   bool   `yaml:"enable" mapstructure:"enable"`
	Addr     string `yaml:"addr" mapstructure:"addr"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

// Config 全局配置
type Config struct {
	App        AppConfig            `yaml:"app" mapstructure:"app"`
	Server     ServerConfig         `yaml:"server" mapstructure:"server"`
	Logger     logger.Config        `yaml:"logger" mapstructure:"logger"`
	I18n       i18n.Config          `yaml:"i18n" mapstructure:"i18n"`
	MySQL      MySQLConfig          `yaml:"mysql" mapstructure:"mysql"`
	Redis      RedisConfig          `yaml:"redis" mapstructure:"redis"`
	Observable observability.Config `yaml:"observable" mapstructure:"observable"`
}

// InitConfig 加载配置文件
func InitConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 热更新监听
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("config file changed: %s\n", e.Name)
		if err := v.Unmarshal(&cfg); err != nil {
			fmt.Printf("failed to reload config: %v\n", err)
		} else {
			fmt.Println("config reloaded successfully")
		}
	})

	return &cfg, nil
}
