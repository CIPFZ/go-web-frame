package plugin

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

// Package plugin -----------------------------
// @file        : redis.go
// @author      : CIPFZ
// @time        : 2025/9/19 17:49
// @description :
// -------------------------------------------

func NewRedis(addr, password string, db int) (*redis.Client, error) {
	opt := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return client, nil
}
