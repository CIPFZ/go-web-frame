package db

import (
	"context"
	"fmt"

	"github.com/CIPFZ/gowebframe/internal/core/config"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func InitRedis(redisCfg config.Redis) (redis.UniversalClient, error) {
	var instance redis.UniversalClient
	var err error
	// 使用集群模式
	if redisCfg.UseCluster {
		instance = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    redisCfg.ClusterAddrs,
			Username: redisCfg.Username,
			Password: redisCfg.Password,
			Protocol: 2,
		})
	} else {
		// 使用单例模式
		instance = redis.NewClient(&redis.Options{
			Addr:     redisCfg.Addr,
			Username: redisCfg.Username,
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
			Protocol: 2,
		})
	}
	// 为客户端添加 Otel 钩子
	if err = redisotel.InstrumentTracing(instance); err != nil {
		return nil, fmt.Errorf("failed to instrument redis with otel: %w", err)
	}

	_, err = instance.Ping(context.Background()).Result()
	return instance, err
}
