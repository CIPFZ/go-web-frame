package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
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
		&sysModel.SysAuthority{},
		&sysModel.SysCasbinRule{},
		&sysModel.SysMenu{},
		&sysModel.SysAuthorityMenu{},
		&sysModel.JwtBlacklist{},
		&sysModel.SysOperationLog{},
		&sysModel.SysUser{},
		&sysModel.SysUserAuthority{},
		&sysModel.SysNotice{},
		&sysModel.SysNoticeReceiver{},

		&model.MetaDynasty{},
		&model.MetaGenre{},
		&model.MetaTag{},
		&model.PoemAuthor{},
		&model.PoemWork{},
		&model.PoemTagRel{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	fmt.Println("AutoMigrate finished successfully!")
}

const defaultConfigPath = "./configs/config.yaml"
