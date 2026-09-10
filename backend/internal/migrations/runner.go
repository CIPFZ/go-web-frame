package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/session"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/seed"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const Latest = "20260910_cms_i18n_v1"
const baselineVersion = "20260910_sessions_policy_bootstrap_v1"

func Check(db *gorm.DB) error {
	var row schemaMigration
	if err := db.Where("name = ?", Latest).First(&row).Error; err != nil {
		return fmt.Errorf("schema is not ready; run migrate before starting server: %w", err)
	}
	return nil
}

// Run pins the advisory lock and all migration work to the same SQL connection.
// MySQL DDL commits implicitly: steps are idempotent, marked only on success.
func Run(ctx context.Context, db *gorm.DB, cfg *config.Config, logger *zap.Logger) error {
	return db.WithContext(ctx).Connection(func(conn *gorm.DB) error {
		// Connection pins ConnPool; reset statement state before each operation.
		conn = conn.Session(&gorm.Session{NewDB: true})
		switch db.Dialector.Name() {
		case "mysql":
			var acquired sql.NullInt64
			if err := conn.Raw("SELECT GET_LOCK('base_frame_migrate', 60)").Scan(&acquired).Error; err != nil {
				return err
			}
			if !acquired.Valid || acquired.Int64 != 1 {
				return fmt.Errorf("migration lock timed out")
			}
			defer conn.WithContext(context.Background()).Exec("SELECT RELEASE_LOCK('base_frame_migrate')")
			return apply(conn, cfg, logger)
		case "postgres":
			// Poll with a bounded context; session lock is released on the pinned connection.
			for {
				var acquired bool
				if err := conn.Raw("SELECT pg_try_advisory_lock(6281739001)").Scan(&acquired).Error; err != nil {
					return err
				}
				if acquired {
					break
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(100 * time.Millisecond):
				}
			}
			defer conn.WithContext(context.Background()).Exec("SELECT pg_advisory_unlock(6281739001)")
			return apply(conn, cfg, logger)
		case "sqlite":
			if err := conn.Exec("CREATE TABLE IF NOT EXISTS sys_migration_lock (id INTEGER PRIMARY KEY, value INTEGER NOT NULL)").Error; err != nil {
				return err
			}
			if err := conn.Exec("INSERT INTO sys_migration_lock (id,value) VALUES (1,0) ON CONFLICT DO NOTHING").Error; err != nil {
				return err
			}
			return conn.Transaction(func(tx *gorm.DB) error {
				if err := tx.Exec("UPDATE sys_migration_lock SET value=value+1 WHERE id=1").Error; err != nil {
					return err
				}
				return apply(tx, cfg, logger)
			})

		default:
			return fmt.Errorf("unsupported migration dialect %s", db.Dialector.Name())
		}
	})
}
func applyBaseline(db *gorm.DB, cfg *config.Config, logger *zap.Logger) error {
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&schemaMigration{}).Where("name = ?", baselineVersion).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := db.AutoMigrate(
		&sysModel.SysApi{}, &sysModel.SysAuthorityApi{}, &sysModel.SysAuthority{}, &sysModel.SysCasbinRule{},
		&sysModel.SysMenu{}, &sysModel.SysAuthorityMenu{}, &sysModel.SysApiToken{}, &sysModel.SysApiTokenApi{},
		&sysModel.JwtBlacklist{}, &sysModel.SysOperationLog{}, &sysModel.SysUser{}, &sysModel.SysUserAuthority{},
		&sysModel.SysNotice{}, &sysModel.SysNoticeReceiver{}, &session.Record{}, &claims.PolicyRevision{},
	); err != nil {
		return err
	}
	if err := db.Model(&sysModel.SysUser{}).Where("status = 2").Update("status", 0).Error; err != nil {
		return err
	}
	if err := cleanupLegacyModules(db, cfg.System.RouterPrefix); err != nil {
		return err
	}
	if err := normalizeCMSBaseline(db); err != nil {
		return err
	}
	if err := reorderCMSMenus(db); err != nil {
		return err
	}
	if err := sanitizeAuditHistory(db); err != nil {
		return err
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&claims.PolicyRevision{ID: 1, Version: 1}).Error; err != nil {
		return err
	}
	if err := seed.Bootstrap(db.Statement.Context, db, cfg, logger); err != nil {
		return err
	}
	if err := db.Model(&sysModel.SysMenu{}).Where("path = ? AND name = ?", "/state", "服务器状态").Update("name", "系统状态").Error; err != nil {
		return err
	}
	return db.Create(&schemaMigration{Name: baselineVersion, AppliedAt: time.Now()}).Error
}

func apply(db *gorm.DB, cfg *config.Config, logger *zap.Logger) error {
	if err := applyBaseline(db, cfg, logger); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&schemaMigration{}).Where("name = ?", Latest).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := db.AutoMigrate(&sysModel.SysMenu{}); err != nil {
		return err
	}
	for key, nameEn := range seed.MenuNamesEn {
		// Preserve user-edited names, English labels, routes, ordering and grants.
		if err := db.Model(&sysModel.SysMenu{}).Where("locale = ? AND name = ? AND (name_en = '' OR name_en IS NULL)", key, seed.MenuNames[key]).Update("name_en", nameEn).Error; err != nil {
			return err
		}
	}
	return db.Create(&schemaMigration{Name: Latest, AppliedAt: time.Now()}).Error
}
