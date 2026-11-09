package migrations

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// userIdentitiesVersion：外部登录提供方的身份关联到本地用户（sys_user_identities）。
const userIdentitiesVersion = "20261109_user_identities_v1"

func applyUserIdentities(db *gorm.DB, _ *config.Config, _ *zap.Logger) error {
	return db.AutoMigrate(&sysModel.SysUserIdentity{})
}
