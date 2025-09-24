package global

import (
	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/pkg/i18n"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Package global -----------------------------
// @file        : global.go
// @author      : CIPFZ
// @time        : 2025/9/20 15:07
// @description :
// -------------------------------------------

type Global struct {
	Config *config.Config
	Logger *zap.Logger
	I18n   *i18n.Service
	DB     *gorm.DB
	Redis  *redis.Client
}

var G = &Global{}
