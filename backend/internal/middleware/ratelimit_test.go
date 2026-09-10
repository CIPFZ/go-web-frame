package middleware

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLimiterStateIsBoundedAndExpires(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1)
	for n := 0; n < 12000; n++ {
		limiter.GetLimiter(fmt.Sprint(n))
	}
	require.Len(t, limiter.ips, 10000)
	limiter.nextSweep = time.Time{}
	for key, e := range limiter.ips {
		e.seen = time.Now().Add(-11 * time.Minute)
		limiter.ips[key] = e
	}
	limiter.GetLimiter("new")
	require.Len(t, limiter.ips, 1)
}
