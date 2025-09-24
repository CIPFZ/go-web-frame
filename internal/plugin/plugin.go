package plugin

// Package plugin -----------------------------
// @file        : plugin.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:50
// @description :
// -------------------------------------------

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	RedisClient *redis.Client
	MySQLClient *gorm.DB
)

func Init(cfg *config.Config) error {
	if cfg.Redis.Enable {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	}
	if cfg.MySQL.Enable {
		db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("mysql init failed: %w", err)
		}
		MySQLClient = db
	}
	return nil
}
