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
	Mysql      MySQL           `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Mongo      Mongo           `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
	Redis      Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	File       FileConfig      `mapstructure:"file" json:"file" yaml:"file"`
	Email      Email           `mapstructure:"email" json:"email" yaml:"email"`
	Captcha    Captcha         `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	Cors       CORS            `mapstructure:"cors" json:"cors" yaml:"cors"`
	Observable Observability   `mapstructure:"observable" json:"observable" yaml:"observable"`
	RateLimit  RateLimitConfig `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`
}

type System struct {
	Name          string `mapstructure:"name" json:"name" yaml:"name"`
	Environment   string `mapstructure:"environment" json:"environment" yaml:"environment"`
	Port          int    `mapstructure:"port" json:"port" yaml:"port"` // 端口值
	RouterPrefix  string `mapstructure:"router_prefix" json:"router_prefix" yaml:"router_prefix"`
	LimitCountIP  int    `mapstructure:"iplimit_count" json:"iplimit_count" yaml:"iplimit_count"`
	LimitTimeIP   int    `mapstructure:"iplimit_time" json:"iplimit_time" yaml:"iplimit_time"`
	DbType        string `mapstructure:"db_type" json:"db_type" yaml:"db_type"`                      // 数据库类型:mysql(默认)|sqlite|postgresql
	UseRedis      bool   `mapstructure:"use_redis" json:"use_redis" yaml:"use_redis"`                // 使用redis
	UseMongo      bool   `mapstructure:"use_mongo" json:"use_mongo" yaml:"use_mongo"`                // 使用mongo
	UseMultipoint bool   `mapstructure:"use_multipoint" json:"use_multipoint" yaml:"use_multipoint"` // 多点登录拦截
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

type Email struct {
	To          string `mapstructure:"to" json:"to" yaml:"to"`                               // 收件人:多个以英文逗号分隔 例：a@qq.com b@qq.com 正式开发中请把此项目作为参数使用
	From        string `mapstructure:"from" json:"from" yaml:"from"`                         // 发件人  你自己要发邮件的邮箱
	Host        string `mapstructure:"host" json:"host" yaml:"host"`                         // 服务器地址 例如 smtp.qq.com  请前往QQ或者你要发邮件的邮箱查看其smtp协议
	Secret      string `mapstructure:"secret" json:"secret" yaml:"secret"`                   // 密钥    用于登录的密钥 最好不要用邮箱密码 去邮箱smtp申请一个用于登录的密钥
	Nickname    string `mapstructure:"nickname" json:"nickname" yaml:"nickname"`             // 昵称    发件人昵称 通常为自己的邮箱
	Port        int    `mapstructure:"port" json:"port" yaml:"port"`                         // 端口     请前往QQ或者你要发邮件的邮箱查看其smtp协议 大多为 465
	IsSSL       bool   `mapstructure:"is_ssl" json:"is_ssl" yaml:"is_ssl"`                   // 是否SSL   是否开启SSL
	IsLoginAuth bool   `mapstructure:"is_loginauth" json:"is_loginauth" yaml:"is_loginauth"` // 是否LoginAuth   是否使用LoginAuth认证方式（适用于IBM、微软邮箱服务器等）
}

type Captcha struct {
	KeyLong            int `mapstructure:"key_long" json:"key_long" yaml:"key_long"`                                     // 验证码长度
	ImgWidth           int `mapstructure:"img_width" json:"img_width" yaml:"img_width"`                                  // 验证码宽度
	ImgHeight          int `mapstructure:"img_height" json:"img_height" yaml:"img_height"`                               // 验证码高度
	OpenCaptcha        int `mapstructure:"open_captcha" json:"open_captcha" yaml:"open_captcha"`                         // 防爆破验证码开启此数，0代表每次登录都需要验证码，其他数字代表错误密码次数，如3代表错误三次后出现验证码
	OpenCaptchaTimeOut int `mapstructure:"open_captcha_timeout" json:"open_captcha_timeout" yaml:"open_captcha_timeout"` // 防爆破验证码超时时间，单位：s(秒)
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
