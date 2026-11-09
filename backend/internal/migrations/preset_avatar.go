package migrations

import (
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// presetAvatarVersion：除了上传头像，也可以从内置头像里选一个。新接口授予能上传头像的角色；已有的头像不变。
const presetAvatarVersion = "20261109_preset_avatar_v1"

func applyPresetAvatar(db *gorm.DB, cfg *config.Config, _ *zap.Logger) error {
	prefix := strings.TrimRight(cfg.System.RouterPrefix, "/") + "/sys/user/"
	if err := ensurePolicyRevision(db); err != nil {
		return err
	}
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		api, err := ensureAPI(tx, sysModel.SysApi{Path: prefix + "avatar/preset", Method: "PUT", ApiGroup: "system-user", Description: "Choose a built-in avatar"})
		if err != nil {
			return err
		}
		roles, err := rolesWithAPI(tx, prefix+"avatar", "POST")
		if err != nil {
			return err
		}
		for _, r := range roles {
			if err := grantAPI(tx, r, api); err != nil {
				return err
			}
		}
		return nil
	})
}
