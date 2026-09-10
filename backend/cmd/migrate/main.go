package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("f", defaultConfigPath, "config file path")
	flag.Parse()

	cfg, _, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	gormDB, err := db.InitDatabase(cfg.Database, logger)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	fmt.Println("Start AutoMigrate...")
	err = gormDB.AutoMigrate(
		&sysModel.SysApi{},
		&sysModel.SysAuthorityApi{},
		&sysModel.SysAuthority{},
		&sysModel.SysCasbinRule{},
		&sysModel.SysMenu{},
		&sysModel.SysAuthorityMenu{},
		&sysModel.SysApiToken{},
		&sysModel.SysApiTokenApi{},
		&sysModel.JwtBlacklist{},
		&sysModel.SysOperationLog{},
		&sysModel.SysUser{},
		&sysModel.SysUserAuthority{},
		&sysModel.SysNotice{},
		&sysModel.SysNoticeReceiver{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := cleanupLegacyModules(gormDB, cfg.System.RouterPrefix); err != nil {
		log.Fatalf("legacy module cleanup failed: %v", err)
	}
	if err := normalizeCMSBaseline(gormDB); err != nil {
		log.Fatalf("CMS baseline migration failed: %v", err)
	}
	fmt.Println("AutoMigrate finished successfully!")
}

const defaultConfigPath = "./configs/config.yaml"
