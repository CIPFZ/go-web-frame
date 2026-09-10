package token

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSharedLeaseAcrossClients(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("requires disposable Redis")
	}
	a := redis.NewClient(&redis.Options{Addr: addr})
	defer a.Close()
	b := redis.NewClient(&redis.Options{Addr: addr})
	defer b.Close()
	ctx, release, err := AcquireRedisLease(context.Background(), a, 123, 1)
	require.NoError(t, err)
	_, _, err = AcquireRedisLease(context.Background(), b, 123, 1)
	require.ErrorIs(t, err, ErrQuota)
	release()
	release()
	require.Error(t, ctx.Err())
	_, releaseB, err := AcquireRedisLease(context.Background(), b, 123, 1)
	require.NoError(t, err)
	releaseB()
}
