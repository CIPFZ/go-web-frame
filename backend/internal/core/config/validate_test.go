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

func TestValidateCOS(t *testing.T) {
	ok := COSConfig{SecretID: "id", SecretKey: "key", Bucket: "files-1250000000", Region: "ap-guangzhou"}
	appID, err := ok.AppID()
	if err != nil || appID != "1250000000" {
		t.Fatalf("appID = %q, %v", appID, err)
	}
	if err := validateCOS(ok); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string]COSConfig{
		"missing key":   {SecretID: "id", Bucket: ok.Bucket, Region: ok.Region},
		"bucket appid":  {SecretID: "id", SecretKey: "key", Bucket: "files", Region: ok.Region},
		"bucket format": {SecretID: "id", SecretKey: "key", Bucket: "Files_1250000000", Region: ok.Region},
		"region":        {SecretID: "id", SecretKey: "key", Bucket: ok.Bucket, Region: "guangzhou"},
	} {
		if validateCOS(bad) == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
