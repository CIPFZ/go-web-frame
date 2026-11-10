package migrations

import (
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
)

// multiRoleVersion：一个用户的所有角色同时生效（权限、菜单取并集），不再切换角色；子角色只能拥有父角色权限的一部分。
const multiRoleVersion = "20261109_multi_role_v1"

// applyMultiRole 删除“切换角色”接口及其授权，并把已有子角色超出父角色的接口与菜单收回。
func applyMultiRole(db *gorm.DB, cfg *config.Config, _ *zap.Logger) error {
	prefix := strings.TrimRight(cfg.System.RouterPrefix, "/")
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		if err := removeAPI(tx, prefix+"/sys/user/switchAuthority", "POST"); err != nil {
			return err
		}
		return repository.PruneChildRoles(tx)
	})
}
