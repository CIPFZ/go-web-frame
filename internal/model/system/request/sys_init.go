package request

import (
	"github.com/CIPFZ/gowebframe/internal/config"
)

type InitDB struct {
	AdminPassword string `json:"adminPassword" binding:"required"`
	DBType        string `json:"dbType"`                    // 数据库类型
	Host          string `json:"host"`                      // 服务器地址
	Port          string `json:"port"`                      // 数据库连接端口
	UserName      string `json:"userName"`                  // 数据库用户名
	Password      string `json:"password"`                  // 数据库密码
	DBName        string `json:"dbName" binding:"required"` // 数据库名
	DBPath        string `json:"dbPath"`                    // sqlite数据库文件路径
	Template      string `json:"template"`                  // postgresql指定template
}

// ToMysqlConfig 转换 config.Mysql
func (i *InitDB) ToMysqlConfig() config.Mysql {
	if i.Host == "" {
		i.Host = "127.0.0.1"
	}
	if i.Port == "" {
		i.Port = "3306"
	}
	return config.Mysql{
		GeneralDB: config.GeneralDB{
			Path:         i.Host,
			Port:         i.Port,
			Dbname:       i.DBName,
			Username:     i.UserName,
			Password:     i.Password,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      "error",
			Config:       "charset=utf8mb4&parseTime=True&loc=Local",
		},
	}
}

// ToPgsqlConfig 转换 config.Pgsql
func (i *InitDB) ToPgsqlConfig() config.Pgsql {
	if i.Host == "" {
		i.Host = "127.0.0.1"
	}
	if i.Port == "" {
		i.Port = "5432"
	}
	return config.Pgsql{
		GeneralDB: config.GeneralDB{
			Path:         i.Host,
			Port:         i.Port,
			Dbname:       i.DBName,
			Username:     i.UserName,
			Password:     i.Password,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      "error",
			Config:       "sslmode=disable TimeZone=Asia/Shanghai",
		},
	}
}
