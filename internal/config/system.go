package config

type System struct {
	Name          string `mapstructure:"name" json:"name" yaml:"name"`
	Environment   string `mapstructure:"environment" json:"environment" yaml:"environment"`
	Port          int    `mapstructure:"port" json:"port" yaml:"port"` // 端口值
	RouterPrefix  string `mapstructure:"router_prefix" json:"router_prefix" yaml:"router_prefix"`
	LimitCountIP  int    `mapstructure:"iplimit_count" json:"iplimit_count" yaml:"iplimit_count"`
	LimitTimeIP   int    `mapstructure:"iplimit_time" json:"iplimit_time" yaml:"iplimit_time"`
	DbType        string `mapstructure:"db_type" json:"db_type" yaml:"db_type"`                         // 数据库类型:mysql(默认)|sqlite|postgresql
	UseRedis      bool   `mapstructure:"use_redis" json:"use_redis" yaml:"use_redis"`                   // 使用redis
	UseMongo      bool   `mapstructure:"use_mongo" json:"use_mongo" yaml:"use_mongo"`                   // 使用mongo
	UseMinio      bool   `mapstructure:"use_minio" json:"use_minio" yaml:"use_minio"`                   // 使用mimio
	UseMultipoint bool   `mapstructure:"use_multipoint" json:"use_multipoint" yaml:"use_multipoint"`    // 多点登录拦截
	UseStrictAuth bool   `mapstructure:"use_strict_auth" json:"use_strict_auth" yaml:"use_strict_auth"` // 使用树形角色分配模式
}
