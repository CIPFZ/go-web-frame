package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateServer(t *testing.T) {
	valid := func() *Config {
		return &Config{System: System{Port: 8080}, JWT: JWT{SigningKey: "01234567890123456789012345678901", Issuer: "cms", ExpiresTime: "7d", BufferTime: "1d"}}
	}
	require.NoError(t, ValidateServer(valid()))
	for name, change := range map[string]func(*Config){"key": func(c *Config) { c.JWT.SigningKey = "short" }, "expiry": func(c *Config) { c.JWT.ExpiresTime = "bad" }, "buffer": func(c *Config) { c.JWT.BufferTime = "8d" }, "proxy": func(c *Config) { c.System.TrustedProxies = []string{"*"} }, "quota": func(c *Config) { c.RateLimit.Enabled = true }, "port": func(c *Config) { c.System.Port = 0 }} {
		t.Run(name, func(t *testing.T) { c := valid(); change(c); require.Error(t, ValidateServer(c)) })
	}
}
