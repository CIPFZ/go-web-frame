package config

import (
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// Config 全局配置
type Config struct {
	System     System          `mapstructure:"system" json:"system" yaml:"system"`
	Logger     Logger          `mapstructure:"logger" json:"logger" yaml:"logger"`
	I18n       I18n            `mapstructure:"i18n" json:"i18n" yaml:"i18n"`
	JWT        JWT             `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Database   Database        `mapstructure:"database" json:"database" yaml:"database"`
	Mysql      MySQL           `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Mongo      Mongo           `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
	Redis      Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	File       FileConfig      `mapstructure:"file" json:"file" yaml:"file"`
	Cors       CORS            `mapstructure:"cors" json:"cors" yaml:"cors"`
	Observable Observability   `mapstructure:"observable" json:"observable" yaml:"observable"`
	RateLimit  RateLimitConfig `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`
}

type Database struct {
	Driver   string   `mapstructure:"driver" json:"driver" yaml:"driver"`
	MySQL    MySQL    `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Postgres Postgres `mapstructure:"postgres" json:"postgres" yaml:"postgres"`
	SQLite   SQLite   `mapstructure:"sqlite" json:"sqlite" yaml:"sqlite"`
}

type System struct {
	AllowRegistration bool     `mapstructure:"allow_registration" json:"allow_registration" yaml:"allow_registration"`
	TrustedProxies    []string `mapstructure:"trusted_proxies" json:"trusted_proxies" yaml:"trusted_proxies"`
	Name              string   `mapstructure:"name" json:"name" yaml:"name"`
	Environment       string   `mapstructure:"environment" json:"environment" yaml:"environment"`
	Port              int      `mapstructure:"port" json:"port" yaml:"port"` // 端口值
	RouterPrefix      string   `mapstructure:"router_prefix" json:"router_prefix" yaml:"router_prefix"`
	UseRedis          bool     `mapstructure:"use_redis" json:"use_redis" yaml:"use_redis"` // 使用redis
	UseMongo          bool     `mapstructure:"use_mongo" json:"use_mongo" yaml:"use_mongo"` // 使用mongo
}

type JWT struct {
	SigningKey  string `mapstructure:"signing_key" json:"signing_key" yaml:"signing_key"`    // jwt签名
	ExpiresTime string `mapstructure:"expires_time" json:"expires_time" yaml:"expires_time"` // 过期时间
	BufferTime  string `mapstructure:"buffer_time" json:"buffer_time" yaml:"buffer_time"`    // 缓冲时间
	Issuer      string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`                   // 签发者
}

type Logger struct {
	Level        string `json:"level" yaml:"level" toml:"level" mapstructure:"level"`                         // "debug","info","warn","error"
	Output       string `json:"output" yaml:"output" toml:"output" mapstructure:"output"`                     // "stdout" | "file" | "both"
	Format       string `json:"format" yaml:"format" toml:"format" mapstructure:"format"`                     // "json" | "console"
	FilePath     string `json:"file_path" yaml:"file_path" toml:"file_path" mapstructure:"file_path"`         // 如果使用 file 或 both
	MaxSizeMB    int    `json:"max_size_mb" yaml:"max_size_mb" toml:"max_size_mb" mapstructure:"max_size_mb"` // lumberjack
	MaxBackups   int    `json:"max_backups" yaml:"max_backups" toml:"max_backups" mapstructure:"max_backups"`
	MaxAgeDays   int    `json:"max_age_days" yaml:"max_age_days" toml:"max_age_days" mapstructure:"max_age_days"`
	Compress     bool   `json:"compress" yaml:"compress" toml:"compress" mapstructure:"compress"`
	EnableCaller bool   `json:"enable_caller" yaml:"enable_caller" toml:"enable_caller" mapstructure:"enable_caller"`
	EnableSample bool   `json:"enable_sample" yaml:"enable_sample" toml:"enable_sample" mapstructure:"enable_sample"` // 是否启用采样，减低高频日志压力
}

// OTELLoggerConfig 日志通过 OTEL 进行发送所需要的配置信息
type OTELLoggerConfig struct {
	LogProvider *sdklog.LoggerProvider
	ServiceName string
	ServiceVer  string
	Environment string
}

// I18n 用于 i18n 初始化的配置
type I18n struct {
	Path string `json:"path" yaml:"path" toml:"path" mapstructure:"path"` // 翻译文件目录
}

type Redis struct {
	Name         string   `mapstructure:"name" json:"name" yaml:"name"`                         // 代表当前实例的名字
	Username     string   `mapstructure:"username" json:"username" yaml:"username"`             // 用户名称
	Addr         string   `mapstructure:"addr" json:"addr" yaml:"addr"`                         // 服务器地址:端口
	Password     string   `mapstructure:"password" json:"password" yaml:"password"`             // 密码
	DB           int      `mapstructure:"db" json:"db" yaml:"db"`                               // 单实例模式下redis的哪个数据库
	UseCluster   bool     `mapstructure:"useCluster" json:"useCluster" yaml:"useCluster"`       // 是否使用集群模式
	ClusterAddrs []string `mapstructure:"clusterAddrs" json:"clusterAddrs" yaml:"clusterAddrs"` // 集群模式下的节点地址列表
}

type FileConfig struct {
	Driver   string      `json:"driver" yaml:"driver" toml:"driver" mapstructure:"driver"`
	MaxMb    int64       `json:"max_mb" yaml:"max_mb" toml:"max_mb" mapstructure:"max_mb"`
	AllowExt []string    `json:"allow_ext" yaml:"allow_ext" toml:"allow_ext" mapstructure:"allow_ext"`
	Local    LocalConfig `json:"local" yaml:"local" toml:"local" mapstructure:"local"`
	Minio    MinioConfig `json:"minio" yaml:"minio" toml:"minio" mapstructure:"minio"`
	COS      COSConfig   `json:"cos" yaml:"cos" toml:"cos" mapstructure:"cos"`
}

type LocalConfig struct {
	Path      string `json:"path" yaml:"path" toml:"path" mapstructure:"path"`
	StorePath string `json:"store_path" yaml:"store_path" toml:"store_path" mapstructure:"store_path"`
}

type MinioConfig struct {
	Endpoint        string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`                            // e.g. localhost:9000
	AccessKey       string `mapstructure:"access_key" json:"access_key" yaml:"access_key"`                      // admin
	SecretKey       string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`                      // password123
	Bucket          string `mapstructure:"bucket" json:"bucket" yaml:"bucket"`                                  // e.g. gowebframe
	UseSSL          bool   `mapstructure:"use_ssl" json:"use_ssl" yaml:"use_ssl"`                               // false
	PreviewBasePath string `mapstructure:"preview_base_path" json:"preview_base_path" yaml:"preview_base_path"` // 访问域名前缀 (http://localhost:9000/bucket)
}

// COSConfig 腾讯云对象存储（driver: cos）。公网/内网域名与 APPID 由 bucket + region 推导。
type COSConfig struct {
	SecretID  string `mapstructure:"secret_id" json:"-" yaml:"secret_id"`
	SecretKey string `mapstructure:"secret_key" json:"-" yaml:"secret_key"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket"`             // <name>-<appid>
	Region    string `mapstructure:"region" json:"region" yaml:"region"`             // 如 ap-guangzhou
	KeyPrefix string `mapstructure:"key_prefix" json:"key_prefix" yaml:"key_prefix"` // 经后端上传的文件放在这个前缀下，可为空
}

type CORS struct {
	Mode      string          `mapstructure:"mode" json:"mode" yaml:"mode"`
	Whitelist []CORSWhitelist `mapstructure:"whitelist" json:"whitelist" yaml:"whitelist"`
}

type CORSWhitelist struct {
	AllowOrigin      string `mapstructure:"allow_origin" json:"allow_origin" yaml:"allow_origin"`
	AllowMethods     string `mapstructure:"allow_methods" json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     string `mapstructure:"allow_headers" json:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders    string `mapstructure:"expose_headers" json:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool   `mapstructure:"allow_credentials" json:"allow_credentials" yaml:"allow_credentials"`
}

type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	QPS     int  `mapstructure:"qps" json:"qps" yaml:"qps"`       // 每秒并发数 (放入桶的速度)
	Burst   int  `mapstructure:"burst" json:"burst" yaml:"burst"` // 突发大小 (桶容量)
}
