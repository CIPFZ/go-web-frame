package config

import (
	"github.com/CIPFZ/gowebframe/pkg/i18n"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/CIPFZ/gowebframe/pkg/observability"
)

// Package config -----------------------------
// @file        : config.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:24
// @description :
// -------------------------------------------

// Config 全局配置
type Config struct {
	System     System               `mapstructure:"system" json:"system" yaml:"system"`
	Logger     logger.Config        `mapstructure:"logger" json:"logger" yaml:"logger"`
	I18n       i18n.Config          `mapstructure:"i18n" json:"i18n" yaml:"i18n"`
	JWT        JWT                  `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Mysql      Mysql                `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Pgsql      Pgsql                `mapstructure:"pgsql" json:"pgsql" yaml:"pgsql"`
	Mongo      Mongo                `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
	Redis      Redis                `mapstructure:"redis" json:"redis" yaml:"redis"`
	Minio      Minio                `mapstructure:"minio" json:"minio" yaml:"minio"`
	Email      Email                `mapstructure:"email" json:"email" yaml:"email"`
	Captcha    Captcha              `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	Cors       CORS                 `mapstructure:"cors" json:"cors" yaml:"cors"`
	Observable observability.Config `mapstructure:"observable" json:"observable" yaml:"observable"`
}
