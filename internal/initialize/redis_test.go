package initialize

import (
	"context"
	"testing"

	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMustInitRedis_Success(t *testing.T) {
	serviceCtx := &svc.ServiceContext{
		Config: &config.Config{
			System: config.System{
				UseRedis: true,
			},
			Redis: config.Redis{
				Name:         "",
				Username:     "admin",
				Addr:         "127.0.0.1:6379",
				Password:     "123456",
				DB:           0,
				UseCluster:   false,
				ClusterAddrs: nil,
			},
		},
		Logger: zap.NewNop(),
	}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("捕获到 panic: %v", r)
		}
	}()
	//
	MustInitRedis(serviceCtx)
	// 验证连接
	count, err := serviceCtx.Redis.Exists(context.Background(), "hello").Result()
	t.Logf("count: %d, err: %v", count, err)
}

func TestMustInitRedis_Failed(t *testing.T) {
	serviceCtx := &svc.ServiceContext{
		Config: &config.Config{
			System: config.System{
				UseRedis: true,
			},
			Redis: config.Redis{
				Name:         "",
				Username:     "admin",
				Addr:         "127.0.0.1:6379",
				Password:     "12345",
				DB:           0,
				UseCluster:   false,
				ClusterAddrs: nil,
			},
		},
		Logger: zap.NewNop(),
	}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("捕获到 panic(测试密码不正确，正常): %v", r)
		}
	}()
	//
	MustInitRedis(serviceCtx)
}

func TestMustInitRedis_NotUse(t *testing.T) {
	serviceCtx := &svc.ServiceContext{
		Config: &config.Config{
			System: config.System{
				UseRedis: false,
			},
		},
		Logger: zap.NewNop(),
	}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("捕获到 panic: %v", r)
		}
	}()
	//
	MustInitRedis(serviceCtx)
	require.Nilf(t, serviceCtx.Redis, "redis client require is nil")
}
