package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func Gorm(serviceCtx *svc.ServiceContext) *gorm.DB {
	return initMysqlDatabase(serviceCtx.Config.Mysql)
}

// GormMysqlByConfig 通过传入配置初始化Mysql数据库
func GormMysqlByConfig(m config.Mysql) *gorm.DB {
	return initMysqlDatabase(m)
}

// initMysqlDatabase (打磨后)
func initMysqlDatabase(m config.Mysql) *gorm.DB {
	if m.Dbname == "" {
		panic("mysql dbname is empty")
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

	// 1. 保持 Open 和 panic 逻辑
	db, err := gorm.Open(mysql.New(mysqlConfig), gormCfg)
	if err != nil {
		panic(fmt.Errorf("failed to open mysql: %w", err))
	}

	// ✨ 2. 关键：将 Otel 插件注册也视为启动检查的一部分
	if err := db.Use(otelgorm.NewPlugin(
		otelgorm.WithDBName(m.Dbname),
		// 自动使用全局 TracerProvider
	)); err != nil {
		panic(fmt.Errorf("failed to use otelgorm plugin: %w", err))
	}

	// 3. (实例设置和连接池设置保持不变)
	db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)

	// 添加 Ping 健康检查 (保留 panic)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		panic(fmt.Errorf("failed to ping mysql: %w", err))
	}
	return db
}
