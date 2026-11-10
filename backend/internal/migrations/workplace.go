package migrations

import (
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// workplaceVersion：每个人可以自己布置工作台（小组件与应用图标），每个角色可以有默认工作台。
// sys_users、sys_authorities 各加 workplace 列；保存接口授予能保存界面设置 / 能修改角色的角色。
const workplaceVersion = "20261109_workplace_desk_v1"

func applyWorkplace(db *gorm.DB, cfg *config.Config, _ *zap.Logger) error {
	for _, m := range []any{&sysModel.SysUser{}, &sysModel.SysAuthority{}} {
		if !db.Migrator().HasColumn(m, "workplace") {
			if err := db.Migrator().AddColumn(m, "Workplace"); err != nil {
				return err
			}
		}
	}
	prefix := strings.TrimRight(cfg.System.RouterPrefix, "/")
	if err := ensurePolicyRevision(db); err != nil {
		return err
	}
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		for _, g := range []struct {
			api      sysModel.SysApi
			like     string // 已经能调用这个接口的角色获得新接口
			likeVerb string
		}{
			{sysModel.SysApi{Path: prefix + "/sys/user/workplace", Method: "PUT", ApiGroup: "system-user", Description: "Update my workplace layout"}, prefix + "/sys/user/ui-config", "PUT"},
			{sysModel.SysApi{Path: prefix + "/sys/authority/workplace", Method: "PUT", ApiGroup: "system-authority", Description: "Set the default workplace of a role"}, prefix + "/sys/authority/updateAuthority", "PUT"},
		} {
			api, err := ensureAPI(tx, g.api)
			if err != nil {
				return err
			}
			roles, err := rolesWithAPI(tx, g.like, g.likeVerb)
			if err != nil {
				return err
			}
			for _, r := range roles {
				if err := grantAPI(tx, r, api); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
