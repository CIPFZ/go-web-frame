package jwt

import (
	"context"
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"go.uber.org/zap"
)

func TestIsBlacklistWithoutRedisReturnsFalse(t *testing.T) {
	j := NewJWT(config.JWT{
		SigningKey:  "test",
		ExpiresTime: "1h",
		BufferTime:  "10m",
		Issuer:      "test",
	}, zap.NewNop(), nil)

	if got := j.IsBlacklist(context.Background(), "token"); got {
		t.Fatal("IsBlacklist() = true, want false when redis is disabled")
	}
}

func TestSetBlacklistWithoutRedisIsNoop(t *testing.T) {
	j := NewJWT(config.JWT{
		SigningKey:  "test",
		ExpiresTime: "1h",
		BufferTime:  "10m",
		Issuer:      "test",
	}, zap.NewNop(), nil)

	if err := j.SetBlacklist(context.Background(), "token", 0); err != nil {
		t.Fatalf("SetBlacklist() error = %v, want nil when redis is disabled", err)
	}
}
