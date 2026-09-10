package token

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var ErrQuota = errors.New("token concurrency quota exceeded")
var leaseAcquire = redis.NewScript(`local t=redis.call('TIME');local now=t[1]*1000+math.floor(t[2]/1000);redis.call('ZREMRANGEBYSCORE',KEYS[1],'-inf',now);if redis.call('ZCARD',KEYS[1])>=tonumber(ARGV[1]) then return 0 end;redis.call('ZADD',KEYS[1],now+30000,ARGV[2]);redis.call('PEXPIRE',KEYS[1],60000);return 1`)
var leaseRenew = redis.NewScript(`if not redis.call('ZSCORE',KEYS[1],ARGV[1]) then return 0 end;local t=redis.call('TIME');local now=t[1]*1000+math.floor(t[2]/1000);redis.call('ZADD',KEYS[1],now+30000,ARGV[1]);redis.call('PEXPIRE',KEYS[1],60000);return 1`)

// AcquireRedisLease shares admission across instances. Handlers MUST respect
// cancellation: lease loss cancels work, and crashed process leases expire.
func AcquireRedisLease(parent context.Context, client redis.UniversalClient, tokenID uint, max int) (context.Context, func(), error) {
	if max < 1 {
		max = 1
	}
	key := fmt.Sprintf("cms:token:lease:{%d}", tokenID)
	id := uuid.NewString()
	check, cancel := context.WithTimeout(parent, 2*time.Second)
	n, err := leaseAcquire.Run(check, client, []string{key}, max, id).Int()
	cancel()
	if err != nil {
		return nil, nil, err
	}
	if n != 1 {
		return nil, nil, ErrQuota
	}
	ctx, stop := context.WithCancel(parent)
	done := make(chan struct{})
	var once sync.Once
	release := func() {
		once.Do(func() {
			stop()
			<-done
			cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = client.ZRem(cleanup, key, id).Err()
		})
	}
	go func() {
		defer close(done)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				check, cancel := context.WithTimeout(ctx, 2*time.Second)
				n, err := leaseRenew.Run(check, client, []string{key}, id).Int()
				cancel()
				if err != nil || n != 1 {
					stop()
					return
				}
			}
		}
	}()
	return ctx, release, nil
}
