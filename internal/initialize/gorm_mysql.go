package initialize

import (
	"time"

	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func Gorm(serviceCtx *svc.ServiceContext) *gorm.DB {
	return initMysqlDatabase(serviceCtx.Config.Mysql)
}

// RegisterTables 数据表注册
func RegisterTables(serviceCtx *svc.ServiceContext) {

	serviceCtx.Logger.Info("register table success")
}

// GormMysqlByConfig 通过传入配置初始化Mysql数据库
func GormMysqlByConfig(m config.Mysql) *gorm.DB {
	return initMysqlDatabase(m)
}

// initMysqlDatabase 初始化Mysql数据库的辅助函数
func initMysqlDatabase(m config.Mysql) *gorm.DB {
	if m.Dbname == "" {
		return nil
	}
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}

	gormCfg := &gorm.Config{
		Logger: logger.New(NewWriter(m), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      m.LogLevel(),
			Colorful:      true,
		}),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   m.Prefix,
			SingularTable: m.Singular,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), gormCfg); err != nil {
		panic(err)
	} else {
		db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		return db
	}
}
