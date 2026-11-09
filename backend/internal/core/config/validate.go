package config

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ValidateServer(c *Config) error {
	if c.System.Port < 1 || c.System.Port > 65535 {
		return fmt.Errorf("system.port must be between 1 and 65535")
	}
	if len(c.JWT.SigningKey) < 32 {
		return fmt.Errorf("jwt.signing_key must contain at least 32 bytes")
	}
	if strings.TrimSpace(c.JWT.Issuer) == "" {
		return fmt.Errorf("jwt.issuer is required")
	}
	expiry, err := configDuration(c.JWT.ExpiresTime)
	if err != nil || expiry <= 0 {
		return fmt.Errorf("jwt.expires_time must be a positive Go duration")
	}
	buffer, err := configDuration(c.JWT.BufferTime)
	if err != nil || buffer < 0 || buffer >= expiry {
		return fmt.Errorf("jwt.buffer_time must be nonnegative and less than expiry")
	}
	if c.RateLimit.Enabled && (c.RateLimit.QPS <= 0 || c.RateLimit.Burst <= 0) {
		return fmt.Errorf("enabled rate_limit requires positive qps and burst")
	}
	for _, proxy := range c.System.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return fmt.Errorf("invalid trusted proxy %q", proxy)
			}
		}
	}
	if c.System.UseRedis && c.Redis.Addr == "" && len(c.Redis.ClusterAddrs) == 0 {
		return fmt.Errorf("redis address is required")
	}
	if c.File.Driver == "cos" {
		return validateCOS(c.File.COS)
	}
	return nil
}

var (
	cosBucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*-([0-9]{6,})$`)
	cosRegionPattern = regexp.MustCompile(`^[a-z]+-[a-z0-9-]+$`)
)

func validateCOS(c COSConfig) error {
	if strings.TrimSpace(c.SecretID) == "" || strings.TrimSpace(c.SecretKey) == "" {
		return fmt.Errorf("file.cos requires secret_id and secret_key")
	}
	if _, err := c.AppID(); err != nil {
		return err
	}
	if !cosRegionPattern.MatchString(c.Region) {
		return fmt.Errorf("file.cos.region %q is invalid", c.Region)
	}
	return nil
}

// AppID 返回桶名后缀中的 APPID（COS 桶名格式为 <name>-<appid>）。
func (c COSConfig) AppID() (string, error) {
	m := cosBucketPattern.FindStringSubmatch(c.Bucket)
	if m == nil {
		return "", fmt.Errorf("file.cos.bucket %q must be <name>-<appid>", c.Bucket)
	}
	return m[1], nil
}

func configDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || n > 365 || n < 0 {
			return 0, fmt.Errorf("invalid day duration")
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
