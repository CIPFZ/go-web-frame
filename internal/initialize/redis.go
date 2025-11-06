package initialize

import (
	"context"

	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func MustInitRedis(serviceCtx *svc.ServiceContext) {
	if !serviceCtx.Config.System.UseRedis {
		return
	}
	redisCfg := serviceCtx.Config.Redis
	// 使用集群模式
	if redisCfg.UseCluster {
		serviceCtx.Redis = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    redisCfg.ClusterAddrs,
			Username: redisCfg.Username,
			Password: redisCfg.Password,
		})
	} else {
		// 使用单例模式
		serviceCtx.Redis = redis.NewClient(&redis.Options{
			Addr:     redisCfg.Addr,
			Username: redisCfg.Username,
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
		})
	}
	pong, err := serviceCtx.Redis.Ping(context.Background()).Result()
	if err != nil {
		serviceCtx.Logger.Error("redis connect ping failed, err:", zap.String("name", redisCfg.Name), zap.Error(err))
		panic("init redis failed, err:" + err.Error())
	}

	serviceCtx.Logger.Info("redis connect ping response:", zap.String("name", redisCfg.Name), zap.String("pong", pong))
}
