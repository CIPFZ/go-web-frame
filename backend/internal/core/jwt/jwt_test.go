package jwt

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
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

func TestParseTokenRequiresExpirationIssuerAndAlgorithm(t *testing.T) {
	cfg := config.JWT{SigningKey: "test-key", ExpiresTime: "1h", BufferTime: "10m", Issuer: "cms"}
	j := NewJWT(cfg, zap.NewNop(), nil)
	valid := j.CreateClaims(dto.BaseClaims{})
	token, err := j.CreateToken(valid)
	require.NoError(t, err)
	_, err = j.ParseToken(token)
	require.NoError(t, err)
	noExpiry := valid
	noExpiry.ExpiresAt = nil
	wrongIssuer := valid
	wrongIssuer.Issuer = "different-service"
	for _, c := range []claims.CustomClaims{noExpiry, wrongIssuer} {
		token, err := j.CreateToken(c)
		require.NoError(t, err)
		_, err = j.ParseToken(token)
		require.Error(t, err)
	}
	token, err = jwtlib.NewWithClaims(jwtlib.SigningMethodHS512, valid).SignedString([]byte(cfg.SigningKey))
	require.NoError(t, err)
	_, err = j.ParseToken(token)
	require.Error(t, err)
}
