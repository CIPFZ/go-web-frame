package copyguard

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

// Option 是一个用于配置 Manager 的函数类型
type Option func(*Config) error

// defaultConfig 创建一个带有合理默认值的配置
func defaultConfig() *Config {
	return &Config{
		BandwidthLimit:        0,                 // 0 表示不限制
		IOPSLimit:             0,                 // 0 表示不限制
		Concurrency:           4,                 // 默认并发数为 4
		CPUThreshold:          0.85,              // CPU 占用率超过 85% 时开始降速
		ReservedDiskSpace:     100 * 1024 * 1024, // 默认预留 100MB 磁盘空间
		MetricsSampleInterval: 2 * time.Second,   // 每 2 秒采集一次指标
		Logger:                zap.NewNop(),      // 默认使用一个无操作的 logger
	}
}

// WithBandwidthLimit 设置全局带宽限制 (单位: bytes/s)
// 值为 0 或负数表示不限制。
func WithBandwidthLimit(limit float64) Option {
	return func(c *Config) error {
		if limit < 0 {
			limit = 0
		}
		c.BandwidthLimit = limit
		return nil
	}
}

// WithIOPSLimit 设置全局 IOPS 限制 (单位: 次/s)
// 值为 0 或负数表示不限制。
func WithIOPSLimit(limit float64) Option {
	return func(c *Config) error {
		if limit < 0 {
			limit = 0
		}
		c.IOPSLimit = limit
		return nil
	}
}

// WithConcurrency 设置拷贝任务的并发数
func WithConcurrency(n int) Option {
	return func(c *Config) error {
		if n <= 0 {
			return errors.New("concurrency must be greater than 0")
		}
		c.Concurrency = n
		return nil
	}
}

// WithCPUThreshold 设置触发动态降速的 CPU 使用率阈值 (0.0 to 1.0)
func WithCPUThreshold(threshold float64) Option {
	return func(c *Config) error {
		if threshold <= 0.0 || threshold > 1.0 {
			return errors.New("cpu threshold must be between 0.0 and 1.0")
		}
		c.CPUThreshold = threshold
		return nil
	}
}

// WithReservedDiskSpace 设置目标路径的预留磁盘空间 (单位: bytes)
func WithReservedDiskSpace(bytes uint64) Option {
	return func(c *Config) error {
		c.ReservedDiskSpace = bytes
		return nil
	}
}

// WithMetricsCallback 设置指标回调函数和采样周期
func WithMetricsCallback(interval time.Duration, callback func(Metrics)) Option {
	return func(c *Config) error {
		if interval <= 0 {
			return errors.New("metrics sample interval must be positive")
		}
		c.MetricsSampleInterval = interval
		c.OnMetrics = callback
		return nil
	}
}

// WithLogger 设置要使用的 zap logger
func WithLogger(logger *zap.Logger) Option {
	return func(c *Config) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		c.Logger = logger
		return nil
	}
}
