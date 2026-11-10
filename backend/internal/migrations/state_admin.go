package migrations

import (
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// stateAdminOnlyVersion：系统状态（Go 版本、迁移版本、依赖状态）只给管理员看。
// 有“系统管理”菜单的角色与超级管理员保留，其余角色收回系统状态的菜单与接口；之后由管理员在角色管理里调整。
const stateAdminOnlyVersion = "20261109_state_admin_only_v1"

func applyStateAdminOnly(db *gorm.DB, cfg *config.Config, _ *zap.Logger) error {
	path := strings.TrimRight(cfg.System.RouterPrefix, "/") + "/sys/system/getServerInfo"
	if err := ensurePolicyRevision(db); err != nil {
		return err
	}
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		var admins []uint
		if err := tx.Model(&sysModel.SysAuthorityMenu{}).Joins("JOIN sys_menus ON sys_menus.id = sys_authority_menus.menu_id").
			Where("sys_menus.path = ?", "/sys").Pluck("sys_authority_menus.authority_id", &admins).Error; err != nil {
			return err
		}
		admins = append(admins, 1)
		names := make([]string, 0, len(admins))
		for _, a := range admins {
			names = append(names, strconv.FormatUint(uint64(a), 10))
		}
		if err := tx.Where("authority_id NOT IN ? AND menu_id IN (?)", admins, tx.Model(&sysModel.SysMenu{}).Select("id").Where("component = ?", "state")).
			Delete(&sysModel.SysAuthorityMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("authority_id NOT IN ? AND api_id IN (?)", admins, tx.Model(&sysModel.SysApi{}).Select("id").Where("path = ? AND method = ?", path, "POST")).
			Delete(&sysModel.SysAuthorityApi{}).Error; err != nil {
			return err
		}
		return tx.Where("ptype = ? AND v0 NOT IN ? AND v1 = ? AND v2 = ?", "p", names, path, "POST").Delete(&sysModel.SysCasbinRule{}).Error
	})
}
