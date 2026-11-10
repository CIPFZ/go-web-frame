package migrations

import (
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// removeBusinessModulesVersion：框架不带业务模块。删除早先版本里 novel、virtualization（含 Agent 联动）的菜单、接口与授权；
// 新装的库本来就没有，这一步什么也不做。业务数据表（如 novel_books）不删，需要时由部署方自行处理。
const removeBusinessModulesVersion = "20261110_remove_business_modules_v1"

var removedModulePaths = []string{"/novel/", "/virtualization/"}

func applyRemoveBusinessModules(db *gorm.DB, cfg *config.Config, _ *zap.Logger) error {
	prefix := strings.TrimRight(cfg.System.RouterPrefix, "/")
	if err := ensurePolicyRevision(db); err != nil {
		return err
	}
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		for _, root := range removedModulePaths {
			var menus []string
			// 子菜单先删（removeMenu 按 path 删除菜单及其授权）
			if err := tx.Unscoped().Model(&sysModel.SysMenu{}).Where("path LIKE ?", root+"%").Order("parent_id DESC").Pluck("path", &menus).Error; err != nil {
				return err
			}
			for _, path := range menus {
				if err := removeMenu(tx, path); err != nil {
					return err
				}
			}
			var apis []sysModel.SysApi
			if err := tx.Unscoped().Where("path LIKE ?", prefix+root+"%").Find(&apis).Error; err != nil {
				return err
			}
			for _, api := range apis {
				if err := removeAPI(tx, api.Path, api.Method); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
