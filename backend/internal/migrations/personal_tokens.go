package migrations

import (
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/seed"
)

// personalTokensVersion：API Token 跟人绑定——每个人管理自己的 Token，权限不超过本人；管理员看所有人的 Token、设置申请限制。
const personalTokensVersion = "20261109_personal_tokens_v1"

// personalTokenAPIs 是每个角色都能调用的“我的 Token”接口（路径相对 RouterPrefix）。
var personalTokenAPIs = []sysModel.SysApi{
	{Path: "/sys/api-token/getApiTokenList", Method: "POST", ApiGroup: "system-api-token", Description: "List my API tokens"},
	{Path: "/sys/api-token/options", Method: "POST", ApiGroup: "system-api-token", Description: "List permissions and limits for my tokens"},
	{Path: "/sys/api-token/detail", Method: "GET", ApiGroup: "system-api-token", Description: "Get my API token"},
	{Path: "/sys/api-token/create", Method: "POST", ApiGroup: "system-api-token", Description: "Create my API token"},
	{Path: "/sys/api-token/update", Method: "PUT", ApiGroup: "system-api-token", Description: "Update my API token"},
	{Path: "/sys/api-token/delete", Method: "DELETE", ApiGroup: "system-api-token", Description: "Delete my API token"},
	{Path: "/sys/api-token/reset", Method: "POST", ApiGroup: "system-api-token", Description: "Reset my API token"},
	{Path: "/sys/api-token/enable", Method: "POST", ApiGroup: "system-api-token", Description: "Enable my API token"},
	{Path: "/sys/api-token/disable", Method: "POST", ApiGroup: "system-api-token", Description: "Disable my API token"},
}

// adminTokenAPIs 是管理所有人 Token 的接口，授予原来就能管理 Token 的角色。
var adminTokenAPIs = []sysModel.SysApi{
	{Path: "/sys/api-token/all", Method: "POST", ApiGroup: "system-api-token", Description: "List everyone's API tokens"},
	{Path: "/sys/api-token/limits/get", Method: "POST", ApiGroup: "system-api-token", Description: "Get token limits"},
	{Path: "/sys/api-token/limits/update", Method: "PUT", ApiGroup: "system-api-token", Description: "Update token limits"},
}

func applyPersonalTokens(db *gorm.DB, cfg *config.Config, _ *zap.Logger) error {
	if err := db.AutoMigrate(&sysModel.SysApiToken{}, &sysModel.SysTokenLimit{}, &sysModel.SysOperationLog{}); err != nil {
		return err
	}
	// Token 必须有主人：没有主人（或主人已删除）的旧 Token 归给最早的超级管理员
	var admins []uint
	if err := db.Table("sys_user_authorities AS ua").Joins("JOIN sys_users AS u ON u.id = ua.user_id AND u.deleted_at IS NULL").
		Where("ua.authority_id = ?", 1).Order("ua.user_id").Limit(1).Pluck("ua.user_id", &admins).Error; err != nil {
		return err
	}
	if len(admins) > 0 {
		alive := db.Model(&sysModel.SysUser{}).Select("id")
		if err := db.Model(&sysModel.SysApiToken{}).Where("created_by = 0 OR created_by NOT IN (?)", alive).
			Update("created_by", admins[0]).Error; err != nil {
			return err
		}
	}

	prefix := strings.TrimRight(cfg.System.RouterPrefix, "/")
	if err := ensurePolicyRevision(db); err != nil {
		return err
	}
	return claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error {
		// 先确定原来能管理 Token 的角色，再给所有人开放“我的 Token”
		managers, err := rolesWithAPI(tx, prefix+"/sys/api-token/getApiTokenList", "POST")
		if err != nil {
			return err
		}
		everyone, err := rolesWithAPI(tx, prefix+"/sys/menu/getMenu", "GET")
		if err != nil {
			return err
		}
		grant := func(apis []sysModel.SysApi, roles []uint) error {
			for _, want := range apis {
				want.Path = prefix + want.Path
				api, err := ensureAPI(tx, want)
				if err != nil {
					return err
				}
				for _, role := range roles {
					if err := grantAPI(tx, role, api); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := grant(adminTokenAPIs, managers); err != nil {
			return err
		}
		if err := grant(personalTokenAPIs, everyone); err != nil {
			return err
		}
		menu, _, err := ensureMenu(tx, sysModel.SysMenu{
			Path: "/account/tokens", Name: seed.MenuNames["menu.account.tokens"], NameEn: seed.MenuNamesEn["menu.account.tokens"],
			Component: "account/tokens", Icon: "KeyOutlined", Locale: "menu.account.tokens", Sort: seed.MenuOrder["/account/tokens"], HideInMenu: true,
		})
		if err != nil {
			return err
		}
		return grantMenu(tx, menu.ID, everyone...)
	})
}
