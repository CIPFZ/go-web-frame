package config

import (
	"fmt"
	"strings"
)

type Mongo struct {
	Coll             string       `json:"coll" yaml:"coll" mapstructure:"coll"`                                           // collection name
	Options          string       `json:"options" yaml:"options" mapstructure:"options"`                                  // mongodb options
	Database         string       `json:"database" yaml:"database" mapstructure:"database"`                               // database name
	Username         string       `json:"username" yaml:"username" mapstructure:"username"`                               // 用户名
	Password         string       `json:"password" yaml:"password" mapstructure:"password"`                               // 密码
	AuthSource       string       `json:"auth_source" yaml:"auth_source" mapstructure:"auth_source"`                      // 验证数据库
	MinPoolSize      uint64       `json:"min_pool_size" yaml:"min_pool_size" mapstructure:"min_pool_size"`                // 最小连接池
	MaxPoolSize      uint64       `json:"max_pool_size" yaml:"max_pool_size" mapstructure:"max_pool_size"`                // 最大连接池
	SocketTimeoutMs  int64        `json:"socket_timeout_ms" yaml:"socket_timeout_ms" mapstructure:"socket_timeout_ms"`    // socket超时时间
	ConnectTimeoutMs int64        `json:"connect_timeout_ms" yaml:"connect_timeout_ms" mapstructure:"connect_timeout_ms"` // 连接超时时间
	IsZap            bool         `json:"is_zap" yaml:"is_zap" mapstructure:"is_zap"`                                     // 是否开启zap日志
	Hosts            []*MongoHost `json:"hosts" yaml:"hosts" mapstructure:"hosts"`                                        // 主机列表
}

type MongoHost struct {
	Host string `json:"host" yaml:"host" mapstructure:"host"` // ip地址
	Port string `json:"port" yaml:"port" mapstructure:"port"` // 端口
}

// Uri .
func (x *Mongo) Uri() string {
	length := len(x.Hosts)
	hosts := make([]string, 0, length)
	for i := 0; i < length; i++ {
		if x.Hosts[i].Host != "" && x.Hosts[i].Port != "" {
			hosts = append(hosts, x.Hosts[i].Host+":"+x.Hosts[i].Port)
		}
	}
	if x.Options != "" {
		return fmt.Sprintf("mongodb://%s/%s?%s", strings.Join(hosts, ","), x.Database, x.Options)
	}
	return fmt.Sprintf("mongodb://%s/%s", strings.Join(hosts, ","), x.Database)
}
