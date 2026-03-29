package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	pluginModel "github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	poetryModel "github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultSeedAdminEnabled      = true
	defaultSeedAdminUsername     = "admin"
	defaultSeedAdminPassword     = "Admin@123456"
	defaultSeedAdminAuthorityID  = uint(1)
	defaultSeedAdminDefaultRoute = "dashboard/workplace"
)

type seedAdminOptions struct {
	Enabled      bool
	Username     string
	Password     string
	AuthorityID  uint
	DefaultRoute string
}

type seedMenu struct {
	Key        string
	ParentKey  string
	Path       string
	Name       string
	Component  string
	Icon       string
	Locale     string
	Access     string
	Target     string
	Sort       int
	HideInMenu bool
}

type seedApi struct {
	Path        string
	Method      string
	ApiGroup    string
	Description string
}

type casbinRuleSeed struct {
	Ptype string `gorm:"column:ptype"`
	V0    string `gorm:"column:v0"`
	V1    string `gorm:"column:v1"`
	V2    string `gorm:"column:v2"`
	V3    string `gorm:"column:v3"`
	V4    string `gorm:"column:v4"`
	V5    string `gorm:"column:v5"`
}

func (casbinRuleSeed) TableName() string {
	return "sys_casbin_rules"
}

func seedAdminIfNeeded(ctx context.Context, serviceCtx *svc.ServiceContext) error {
	opts := loadSeedAdminOptions(serviceCtx)
	if !opts.Enabled {
		serviceCtx.Logger.Info("seed admin disabled")
		return nil
	}

	return serviceCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureAuthorities(tx, opts); err != nil {
			return err
		}

		menuIDs, err := ensureBaseMenus(tx)
		if err != nil {
			return err
		}
		if err := bindAuthorityMenus(tx, menuIDs, opts.AuthorityID); err != nil {
			return err
		}

		apiIDs, err := ensureBaseApis(tx)
		if err != nil {
			return err
		}
		if err := bindAuthorityApis(tx, apiIDs, opts.AuthorityID); err != nil {
			return err
		}

		if err := ensureCasbinPolicies(tx, apiIDs, opts.AuthorityID); err != nil {
			return err
		}

		if err := ensureAdminUser(tx, opts); err != nil {
			return err
		}
		if err := ensurePoetryBaseData(tx); err != nil {
			return err
		}
		if err := ensurePluginBaseData(tx); err != nil {
			return err
		}

		serviceCtx.Logger.Info("seed base system data finished",
			zap.String("username", opts.Username),
			zap.Uint("authorityId", opts.AuthorityID),
		)
		return nil
	})
}

func loadSeedAdminOptions(serviceCtx *svc.ServiceContext) seedAdminOptions {
	env := strings.ToLower(strings.TrimSpace(serviceCtx.Config.System.Environment))
	isDevLike := env == "dev" || env == "development" || env == "local"

	opts := seedAdminOptions{
		Enabled:      defaultSeedAdminEnabled && isDevLike,
		Username:     defaultSeedAdminUsername,
		Password:     defaultSeedAdminPassword,
		AuthorityID:  defaultSeedAdminAuthorityID,
		DefaultRoute: defaultSeedAdminDefaultRoute,
	}

	if v := strings.TrimSpace(os.Getenv("SEED_ADMIN_ENABLED")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			opts.Enabled = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("SEED_ADMIN_USERNAME")); v != "" {
		opts.Username = v
	}
	if v := strings.TrimSpace(os.Getenv("SEED_ADMIN_PASSWORD")); v != "" {
		opts.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("SEED_ADMIN_AUTHORITY_ID")); v != "" {
		if parsed, err := strconv.ParseUint(v, 10, 32); err == nil {
			opts.AuthorityID = uint(parsed)
		}
	}
	if v := strings.TrimSpace(os.Getenv("SEED_ADMIN_DEFAULT_ROUTER")); v != "" {
		opts.DefaultRoute = strings.Trim(v, "/")
	}
	return opts
}

func ensureAuthorities(tx *gorm.DB, opts seedAdminOptions) error {
	authorities := []model.SysAuthority{
		{AuthorityId: opts.AuthorityID, AuthorityName: "Administrator", ParentId: 0, DefaultRouter: opts.DefaultRoute},
		{AuthorityId: 888, AuthorityName: "CommonUser", ParentId: 0, DefaultRouter: "dashboard/workplace"},
		{AuthorityId: 9528, AuthorityName: "TestUser", ParentId: 0, DefaultRouter: "dashboard/workplace"},
		{AuthorityId: 8881, AuthorityName: "CommonUserChild", ParentId: 888, DefaultRouter: "dashboard/workplace"},
		{AuthorityId: 10010, AuthorityName: "PluginRequester", ParentId: 0, DefaultRouter: "plugin/center"},
		{AuthorityId: 10013, AuthorityName: "PluginReviewer", ParentId: 0, DefaultRouter: "plugin/center"},
		{AuthorityId: 10014, AuthorityName: "PluginPublisher", ParentId: 0, DefaultRouter: "plugin/center"},
	}

	seen := make(map[uint]struct{}, len(authorities))
	for _, auth := range authorities {
		if _, ok := seen[auth.AuthorityId]; ok {
			continue
		}
		seen[auth.AuthorityId] = struct{}{}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&auth).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureBaseMenus(tx *gorm.DB) (map[string]uint, error) {
	menus := []seedMenu{
		{Key: "dashboard", Path: "/dashboard/workplace", Name: "menu.dashboard.workplace", Component: "dashboard/workplace", Icon: "DashboardOutlined", Sort: 1, Locale: "menu.dashboard.workplace"},
		{Key: "state", Path: "/state", Name: "menu.state", Component: "state", Icon: "CloudServerOutlined", Sort: 2, Locale: "menu.state"},
		{Key: "about", Path: "/about", Name: "menu.about", Component: "about", Icon: "InfoCircleOutlined", Sort: 3, Locale: "menu.about"},
		{Key: "sys_root", Path: "/sys", Name: "menu.system", Component: "components/RouterLayout", Icon: "SettingOutlined", Sort: 10, Locale: "menu.system"},
		{Key: "sys_user", ParentKey: "sys_root", Path: "/sys/user", Name: "menu.system.user", Component: "sys/user", Icon: "UserOutlined", Sort: 1, Locale: "menu.system.user"},
		{Key: "sys_authority", ParentKey: "sys_root", Path: "/sys/authority", Name: "menu.system.authority", Component: "sys/authority", Icon: "TeamOutlined", Sort: 2, Locale: "menu.system.authority"},
		{Key: "sys_menu", ParentKey: "sys_root", Path: "/sys/menu", Name: "menu.system.menu", Component: "sys/menu", Icon: "MenuOutlined", Sort: 3, Locale: "menu.system.menu"},
		{Key: "sys_api", ParentKey: "sys_root", Path: "/sys/api", Name: "menu.system.api", Component: "sys/api", Icon: "ApiOutlined", Sort: 4, Locale: "menu.system.api"},
		{Key: "sys_operation", ParentKey: "sys_root", Path: "/sys/operation", Name: "menu.system.operation", Component: "sys/operation", Icon: "HistoryOutlined", Sort: 5, Locale: "menu.system.operation"},
		{Key: "sys_notice", ParentKey: "sys_root", Path: "/sys/notice", Name: "menu.system.notice", Component: "sys/notice", Icon: "NotificationOutlined", Sort: 6, Locale: "menu.system.notice"},
		{Key: "plugin_root", Path: "/plugin", Name: "menu.plugin", Component: "components/RouterLayout", Icon: "AppstoreOutlined", Sort: 15, Locale: "menu.plugin"},
		{Key: "plugin_center", ParentKey: "plugin_root", Path: "/plugin/center", Name: "menu.plugin.center", Component: "plugin/center", Icon: "DeploymentUnitOutlined", Sort: 1, Locale: "menu.plugin.center"},
		{Key: "plugin_review_workbench", ParentKey: "plugin_root", Path: "/plugin/review-workbench", Name: "menu.plugin.review-workbench", Component: "plugin/review-workbench", Icon: "AuditOutlined", Sort: 2, Locale: "menu.plugin.review-workbench"},
		{Key: "plugin_publish_workbench", ParentKey: "plugin_root", Path: "/plugin/publish-workbench", Name: "menu.plugin.publish-workbench", Component: "plugin/publish-workbench", Icon: "RocketOutlined", Sort: 3, Locale: "menu.plugin.publish-workbench"},
		{Key: "plugin_project_detail", ParentKey: "plugin_root", Path: "/plugin/project/:id", Name: "menu.plugin.center", Component: "plugin/project", Icon: "DeploymentUnitOutlined", Sort: 99, Locale: "menu.plugin.center", HideInMenu: true},
		{Key: "poetry_root", Path: "/poetry", Name: "menu.poetry", Component: "components/RouterLayout", Icon: "BookOutlined", Sort: 20, Locale: "menu.poetry"},
		{Key: "poetry_dynasty", ParentKey: "poetry_root", Path: "/poetry/dynasty", Name: "menu.poetry.dynasty", Component: "poetry/dynasty", Icon: "AppstoreOutlined", Sort: 1, Locale: "menu.poetry.dynasty"},
		{Key: "poetry_genre", ParentKey: "poetry_root", Path: "/poetry/genre", Name: "menu.poetry.genre", Component: "poetry/genre", Icon: "TagsOutlined", Sort: 2, Locale: "menu.poetry.genre"},
		{Key: "poetry_author", ParentKey: "poetry_root", Path: "/poetry/author", Name: "menu.poetry.author", Component: "poetry/author", Icon: "UsergroupAddOutlined", Sort: 3, Locale: "menu.poetry.author"},
		{Key: "poetry_poem", ParentKey: "poetry_root", Path: "/poetry/poem", Name: "menu.poetry.poem", Component: "poetry/poem", Icon: "ReadOutlined", Sort: 4, Locale: "menu.poetry.poem"},
		{Key: "account_settings", Path: "/account/settings", Name: "menu.account.settings", Component: "user/info", Icon: "ProfileOutlined", Sort: 99, Locale: "menu.account.settings", HideInMenu: true},
	}

	menuIDs := make(map[string]uint, len(menus))
	for _, item := range menus {
		parentID := uint(0)
		if item.ParentKey != "" {
			pid, ok := menuIDs[item.ParentKey]
			if !ok {
				return nil, errors.New("seed menu parent not found: " + item.ParentKey)
			}
			parentID = pid
		}

		var menu model.SysMenu
		err := tx.Where("path = ?", item.Path).First(&menu).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			menu = model.SysMenu{
				ParentId:   parentID,
				Path:       item.Path,
				Name:       item.Name,
				Component:  item.Component,
				Access:     item.Access,
				Target:     item.Target,
				Locale:     item.Locale,
				Sort:       item.Sort,
				Icon:       item.Icon,
				HideInMenu: item.HideInMenu,
			}
			if err := tx.Create(&menu).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			updates := map[string]interface{}{
				"parent_id":    parentID,
				"name":         item.Name,
				"component":    item.Component,
				"access":       item.Access,
				"target":       item.Target,
				"locale":       item.Locale,
				"sort":         item.Sort,
				"icon":         item.Icon,
				"hide_in_menu": item.HideInMenu,
			}
			if err := tx.Model(&menu).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		menuIDs[item.Key] = menu.ID
	}
	return menuIDs, nil
}

func bindAuthorityMenus(tx *gorm.DB, menuIDs map[string]uint, adminAuthorityID uint) error {
	roleMenuKeys := map[uint][]string{
		adminAuthorityID: {"dashboard", "state", "about", "sys_root", "sys_user", "sys_authority", "sys_menu", "sys_api", "sys_operation", "sys_notice", "plugin_root", "plugin_center", "plugin_review_workbench", "plugin_publish_workbench", "plugin_project_detail", "poetry_root", "poetry_dynasty", "poetry_genre", "poetry_author", "poetry_poem", "account_settings"},
		9528:             {"dashboard", "state", "about", "sys_root", "sys_user", "sys_authority", "sys_menu", "sys_api", "sys_operation", "sys_notice", "plugin_root", "plugin_center", "plugin_review_workbench", "plugin_publish_workbench", "plugin_project_detail", "poetry_root", "poetry_dynasty", "poetry_genre", "poetry_author", "poetry_poem", "account_settings"},
		888:              {"dashboard", "state", "about", "poetry_root", "poetry_dynasty", "poetry_genre", "poetry_author", "poetry_poem", "account_settings"},
		8881:             {"dashboard", "about", "account_settings"},
		10010:            {"dashboard", "state", "about", "plugin_root", "plugin_center", "plugin_project_detail", "account_settings"},
		10013:            {"dashboard", "state", "about", "plugin_root", "plugin_review_workbench", "plugin_project_detail", "account_settings"},
		10014:            {"dashboard", "state", "about", "plugin_root", "plugin_publish_workbench", "plugin_project_detail", "account_settings"},
	}

	pluginMenuKeys := []string{"plugin_root", "plugin_center", "plugin_review_workbench", "plugin_publish_workbench", "plugin_project_detail"}
	pluginMenuIDs := make([]uint, 0, len(pluginMenuKeys))
	for _, key := range pluginMenuKeys {
		if menuID, ok := menuIDs[key]; ok {
			pluginMenuIDs = append(pluginMenuIDs, menuID)
		}
	}
	if len(pluginMenuIDs) > 0 {
		targetAuthorities := []uint{adminAuthorityID, 9528, 10010, 10013, 10014}
		if err := tx.Where("authority_id IN ? AND menu_id IN ?", targetAuthorities, pluginMenuIDs).Delete(&model.SysAuthorityMenu{}).Error; err != nil {
			return err
		}
	}

	relations := make([]model.SysAuthorityMenu, 0)
	for authorityID, keys := range roleMenuKeys {
		for _, key := range keys {
			menuID, ok := menuIDs[key]
			if !ok {
				continue
			}
			relations = append(relations, model.SysAuthorityMenu{AuthorityId: authorityID, MenuId: menuID})
		}
	}
	if len(relations) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&relations).Error
}

func ensureBaseApis(tx *gorm.DB) (map[string]uint, error) {
	apis := []seedApi{
		{Path: "/api/v1/sys/user/getSelfInfo", Method: "GET", ApiGroup: "system-user", Description: "Get current user"},
		{Path: "/api/v1/sys/user/getUserList", Method: "POST", ApiGroup: "system-user", Description: "Get user list"},
		{Path: "/api/v1/sys/user/logout", Method: "POST", ApiGroup: "system-user", Description: "Logout"},
		{Path: "/api/v1/sys/user/info", Method: "PUT", ApiGroup: "system-user", Description: "Update self info"},
		{Path: "/api/v1/sys/user/ui-config", Method: "PUT", ApiGroup: "system-user", Description: "Update UI config"},
		{Path: "/api/v1/sys/user/avatar", Method: "POST", ApiGroup: "system-user", Description: "Upload avatar"},
		{Path: "/api/v1/sys/user/switchAuthority", Method: "POST", ApiGroup: "system-user", Description: "Switch authority"},
		{Path: "/api/v1/sys/user/addUser", Method: "POST", ApiGroup: "system-user", Description: "Add user"},
		{Path: "/api/v1/sys/user/updateUser", Method: "PUT", ApiGroup: "system-user", Description: "Update user"},
		{Path: "/api/v1/sys/user/deleteUser", Method: "DELETE", ApiGroup: "system-user", Description: "Delete user"},
		{Path: "/api/v1/sys/user/resetPassword", Method: "POST", ApiGroup: "system-user", Description: "Reset password"},

		{Path: "/api/v1/sys/menu/getMenu", Method: "GET", ApiGroup: "system-menu", Description: "Get current menu"},
		{Path: "/api/v1/sys/menu/getMenuList", Method: "POST", ApiGroup: "system-menu", Description: "Get menu list"},
		{Path: "/api/v1/sys/menu/getMenuAuthority", Method: "POST", ApiGroup: "system-menu", Description: "Get authority menus"},
		{Path: "/api/v1/sys/menu/addBaseMenu", Method: "POST", ApiGroup: "system-menu", Description: "Create menu"},
		{Path: "/api/v1/sys/menu/updateBaseMenu", Method: "PUT", ApiGroup: "system-menu", Description: "Update menu"},
		{Path: "/api/v1/sys/menu/deleteBaseMenu", Method: "DELETE", ApiGroup: "system-menu", Description: "Delete menu"},

		{Path: "/api/v1/sys/authority/getAuthorityList", Method: "POST", ApiGroup: "system-authority", Description: "Get authority list"},
		{Path: "/api/v1/sys/authority/createAuthority", Method: "POST", ApiGroup: "system-authority", Description: "Create authority"},
		{Path: "/api/v1/sys/authority/updateAuthority", Method: "PUT", ApiGroup: "system-authority", Description: "Update authority"},
		{Path: "/api/v1/sys/authority/deleteAuthority", Method: "DELETE", ApiGroup: "system-authority", Description: "Delete authority"},
		{Path: "/api/v1/sys/authority/setAuthorityMenus", Method: "POST", ApiGroup: "system-authority", Description: "Set authority menus"},

		{Path: "/api/v1/sys/api/getApiList", Method: "POST", ApiGroup: "system-api", Description: "Get API list"},
		{Path: "/api/v1/sys/api/createApi", Method: "POST", ApiGroup: "system-api", Description: "Create API"},
		{Path: "/api/v1/sys/api/updateApi", Method: "PUT", ApiGroup: "system-api", Description: "Update API"},
		{Path: "/api/v1/sys/api/deleteApi", Method: "DELETE", ApiGroup: "system-api", Description: "Delete API"},

		{Path: "/api/v1/sys/casbin/getPolicyPathByAuthorityId", Method: "POST", ApiGroup: "system-casbin", Description: "Get casbin policy list"},
		{Path: "/api/v1/sys/casbin/updateCasbin", Method: "POST", ApiGroup: "system-casbin", Description: "Update casbin policy"},

		{Path: "/api/v1/sys/operationLog/getOperationLogList", Method: "POST", ApiGroup: "system-operation", Description: "Get operation logs"},
		{Path: "/api/v1/sys/operationLog/deleteOperationLogByIds", Method: "DELETE", ApiGroup: "system-operation", Description: "Delete operation logs"},
		{Path: "/api/v1/sys/file/upload", Method: "POST", ApiGroup: "system-file", Description: "Upload file"},
		{Path: "/api/v1/sys/system/getServerInfo", Method: "POST", ApiGroup: "system-state", Description: "Get server state"},
		{Path: "/api/v1/sys/notice/createNotice", Method: "POST", ApiGroup: "system-notice", Description: "Create notice"},
		{Path: "/api/v1/sys/notice/getNoticeList", Method: "POST", ApiGroup: "system-notice", Description: "Get notice list"},
		{Path: "/api/v1/sys/notice/getMyNotices", Method: "GET", ApiGroup: "system-notice", Description: "Get my notices"},
		{Path: "/api/v1/sys/notice/markRead", Method: "POST", ApiGroup: "system-notice", Description: "Mark notice as read"},
		{Path: "/api/v1/plugin/plugin/getPluginList", Method: "POST", ApiGroup: "plugin", Description: "Get plugin list"},
		{Path: "/api/v1/plugin/plugin/getPluginOverview", Method: "GET", ApiGroup: "plugin", Description: "Get plugin overview"},
		{Path: "/api/v1/plugin/plugin/createPlugin", Method: "POST", ApiGroup: "plugin", Description: "Create plugin"},
		{Path: "/api/v1/plugin/plugin/updatePlugin", Method: "PUT", ApiGroup: "plugin", Description: "Update plugin"},
		{Path: "/api/v1/plugin/plugin/getProjectDetail", Method: "POST", ApiGroup: "plugin", Description: "Get project detail"},
		{Path: "/api/v1/plugin/release/getReleaseList", Method: "POST", ApiGroup: "plugin", Description: "Get release ticket list"},
		{Path: "/api/v1/plugin/release/getReleaseDetail", Method: "POST", ApiGroup: "plugin", Description: "Get release ticket detail"},
		{Path: "/api/v1/plugin/release/createRelease", Method: "POST", ApiGroup: "plugin", Description: "Create release ticket"},
		{Path: "/api/v1/plugin/release/updateRelease", Method: "PUT", ApiGroup: "plugin", Description: "Update release ticket"},
		{Path: "/api/v1/plugin/release/transition", Method: "POST", ApiGroup: "plugin", Description: "Transit release ticket"},
		{Path: "/api/v1/plugin/public/getPublishedPluginList", Method: "POST", ApiGroup: "plugin-public", Description: "Get published plugin list"},
		{Path: "/api/v1/plugin/public/getPublishedPluginDetail", Method: "POST", ApiGroup: "plugin-public", Description: "Get published plugin detail"},

		{Path: "/api/v1/poetry/dynasty", Method: "POST", ApiGroup: "poetry", Description: "Create dynasty"},
		{Path: "/api/v1/poetry/dynasty/:id", Method: "PUT", ApiGroup: "poetry", Description: "Update dynasty"},
		{Path: "/api/v1/poetry/dynasty/:id", Method: "DELETE", ApiGroup: "poetry", Description: "Delete dynasty"},
		{Path: "/api/v1/poetry/dynasty/list", Method: "GET", ApiGroup: "poetry", Description: "List dynasty"},
		{Path: "/api/v1/poetry/dynasty/all", Method: "GET", ApiGroup: "poetry", Description: "All dynasties"},

		{Path: "/api/v1/poetry/genre", Method: "POST", ApiGroup: "poetry", Description: "Create genre"},
		{Path: "/api/v1/poetry/genre/:id", Method: "PUT", ApiGroup: "poetry", Description: "Update genre"},
		{Path: "/api/v1/poetry/genre/:id", Method: "DELETE", ApiGroup: "poetry", Description: "Delete genre"},
		{Path: "/api/v1/poetry/genre/list", Method: "GET", ApiGroup: "poetry", Description: "List genre"},
		{Path: "/api/v1/poetry/genre/all", Method: "GET", ApiGroup: "poetry", Description: "All genres"},

		{Path: "/api/v1/poetry/author", Method: "POST", ApiGroup: "poetry", Description: "Create author"},
		{Path: "/api/v1/poetry/author/:id", Method: "PUT", ApiGroup: "poetry", Description: "Update author"},
		{Path: "/api/v1/poetry/author/:id", Method: "DELETE", ApiGroup: "poetry", Description: "Delete author"},
		{Path: "/api/v1/poetry/author/list", Method: "GET", ApiGroup: "poetry", Description: "List author"},
		{Path: "/api/v1/poetry/author/:id", Method: "GET", ApiGroup: "poetry", Description: "Author detail"},
		{Path: "/api/v1/poetry/author/avatar", Method: "POST", ApiGroup: "poetry", Description: "Upload author avatar"},

		{Path: "/api/v1/poetry/poem", Method: "POST", ApiGroup: "poetry", Description: "Create poem"},
		{Path: "/api/v1/poetry/poem/:id", Method: "PUT", ApiGroup: "poetry", Description: "Update poem"},
		{Path: "/api/v1/poetry/poem/:id", Method: "DELETE", ApiGroup: "poetry", Description: "Delete poem"},
		{Path: "/api/v1/poetry/poem/list", Method: "GET", ApiGroup: "poetry", Description: "List poem"},
		{Path: "/api/v1/poetry/poem/:id", Method: "GET", ApiGroup: "poetry", Description: "Poem detail"},
	}

	apiIDs := make(map[string]uint, len(apis))
	for _, item := range apis {
		sign := apiSign(item.Method, item.Path)

		var api model.SysApi
		err := tx.Where("path = ? AND method = ?", item.Path, item.Method).First(&api).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			api = model.SysApi{Path: item.Path, Method: item.Method, ApiGroup: item.ApiGroup, Description: item.Description}
			if err := tx.Create(&api).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		apiIDs[sign] = api.ID
	}
	return apiIDs, nil
}

func bindAuthorityApis(tx *gorm.DB, apiIDs map[string]uint, adminAuthorityID uint) error {
	fullAccess := []string{
		apiSign("GET", "/api/v1/sys/user/getSelfInfo"),
		apiSign("POST", "/api/v1/sys/user/getUserList"),
		apiSign("POST", "/api/v1/sys/user/logout"),
		apiSign("PUT", "/api/v1/sys/user/info"),
		apiSign("PUT", "/api/v1/sys/user/ui-config"),
		apiSign("POST", "/api/v1/sys/user/avatar"),
		apiSign("POST", "/api/v1/sys/user/switchAuthority"),
		apiSign("POST", "/api/v1/sys/user/addUser"),
		apiSign("PUT", "/api/v1/sys/user/updateUser"),
		apiSign("DELETE", "/api/v1/sys/user/deleteUser"),
		apiSign("POST", "/api/v1/sys/user/resetPassword"),
		apiSign("GET", "/api/v1/sys/menu/getMenu"),
		apiSign("POST", "/api/v1/sys/menu/getMenuList"),
		apiSign("POST", "/api/v1/sys/menu/getMenuAuthority"),
		apiSign("POST", "/api/v1/sys/menu/addBaseMenu"),
		apiSign("PUT", "/api/v1/sys/menu/updateBaseMenu"),
		apiSign("DELETE", "/api/v1/sys/menu/deleteBaseMenu"),
		apiSign("POST", "/api/v1/sys/authority/getAuthorityList"),
		apiSign("POST", "/api/v1/sys/authority/createAuthority"),
		apiSign("PUT", "/api/v1/sys/authority/updateAuthority"),
		apiSign("DELETE", "/api/v1/sys/authority/deleteAuthority"),
		apiSign("POST", "/api/v1/sys/authority/setAuthorityMenus"),
		apiSign("POST", "/api/v1/sys/api/getApiList"),
		apiSign("POST", "/api/v1/sys/api/createApi"),
		apiSign("PUT", "/api/v1/sys/api/updateApi"),
		apiSign("DELETE", "/api/v1/sys/api/deleteApi"),
		apiSign("POST", "/api/v1/sys/casbin/getPolicyPathByAuthorityId"),
		apiSign("POST", "/api/v1/sys/casbin/updateCasbin"),
		apiSign("POST", "/api/v1/sys/operationLog/getOperationLogList"),
		apiSign("DELETE", "/api/v1/sys/operationLog/deleteOperationLogByIds"),
		apiSign("POST", "/api/v1/sys/file/upload"),
		apiSign("POST", "/api/v1/sys/system/getServerInfo"),
		apiSign("POST", "/api/v1/sys/notice/createNotice"),
		apiSign("POST", "/api/v1/sys/notice/getNoticeList"),
		apiSign("GET", "/api/v1/sys/notice/getMyNotices"),
		apiSign("POST", "/api/v1/sys/notice/markRead"),
		apiSign("POST", "/api/v1/plugin/plugin/getPluginList"),
		apiSign("POST", "/api/v1/plugin/plugin/getProjectDetail"),
		apiSign("POST", "/api/v1/plugin/plugin/createPlugin"),
		apiSign("PUT", "/api/v1/plugin/plugin/updatePlugin"),
		apiSign("POST", "/api/v1/plugin/release/getReleaseList"),
		apiSign("POST", "/api/v1/plugin/release/getReleaseDetail"),
		apiSign("POST", "/api/v1/plugin/release/createRelease"),
		apiSign("PUT", "/api/v1/plugin/release/updateRelease"),
		apiSign("POST", "/api/v1/plugin/release/transition"),
		apiSign("POST", "/api/v1/poetry/dynasty"),
		apiSign("PUT", "/api/v1/poetry/dynasty/:id"),
		apiSign("DELETE", "/api/v1/poetry/dynasty/:id"),
		apiSign("GET", "/api/v1/poetry/dynasty/list"),
		apiSign("GET", "/api/v1/poetry/dynasty/all"),
		apiSign("POST", "/api/v1/poetry/genre"),
		apiSign("PUT", "/api/v1/poetry/genre/:id"),
		apiSign("DELETE", "/api/v1/poetry/genre/:id"),
		apiSign("GET", "/api/v1/poetry/genre/list"),
		apiSign("GET", "/api/v1/poetry/genre/all"),
		apiSign("POST", "/api/v1/poetry/author"),
		apiSign("PUT", "/api/v1/poetry/author/:id"),
		apiSign("DELETE", "/api/v1/poetry/author/:id"),
		apiSign("GET", "/api/v1/poetry/author/list"),
		apiSign("GET", "/api/v1/poetry/author/:id"),
		apiSign("POST", "/api/v1/poetry/author/avatar"),
		apiSign("POST", "/api/v1/poetry/poem"),
		apiSign("PUT", "/api/v1/poetry/poem/:id"),
		apiSign("DELETE", "/api/v1/poetry/poem/:id"),
		apiSign("GET", "/api/v1/poetry/poem/list"),
		apiSign("GET", "/api/v1/poetry/poem/:id"),
	}
	basicAccess := []string{
		apiSign("GET", "/api/v1/sys/user/getSelfInfo"),
		apiSign("POST", "/api/v1/sys/user/logout"),
		apiSign("PUT", "/api/v1/sys/user/info"),
		apiSign("PUT", "/api/v1/sys/user/ui-config"),
		apiSign("POST", "/api/v1/sys/user/avatar"),
		apiSign("GET", "/api/v1/sys/menu/getMenu"),
		apiSign("POST", "/api/v1/sys/system/getServerInfo"),
		apiSign("GET", "/api/v1/sys/notice/getMyNotices"),
		apiSign("POST", "/api/v1/sys/notice/markRead"),
		apiSign("GET", "/api/v1/poetry/dynasty/list"),
		apiSign("GET", "/api/v1/poetry/dynasty/all"),
		apiSign("GET", "/api/v1/poetry/genre/list"),
		apiSign("GET", "/api/v1/poetry/genre/all"),
		apiSign("GET", "/api/v1/poetry/author/list"),
		apiSign("GET", "/api/v1/poetry/author/:id"),
		apiSign("GET", "/api/v1/poetry/poem/list"),
		apiSign("GET", "/api/v1/poetry/poem/:id"),
	}

	roleAccess := map[uint][]string{
		adminAuthorityID: fullAccess,
		9528:             fullAccess,
		888:              basicAccess,
		8881:             basicAccess,
		10010:            append(basicAccess, pluginOnlyAccess()...),
		10013:            append(basicAccess, pluginOnlyAccess()...),
		10014:            append(basicAccess, pluginOnlyAccess()...),
	}

	relations := make([]model.SysAuthorityApi, 0)
	for authorityID, signs := range roleAccess {
		for _, sign := range signs {
			apiID, ok := apiIDs[sign]
			if !ok {
				continue
			}
			relations = append(relations, model.SysAuthorityApi{AuthorityId: authorityID, ApiId: apiID})
		}
	}
	if len(relations) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&relations).Error
}

func ensureCasbinPolicies(tx *gorm.DB, apiIDs map[string]uint, adminAuthorityID uint) error {
	_ = apiIDs
	fullAccess := [][]string{
		{"GET", "/api/v1/sys/user/getSelfInfo"},
		{"POST", "/api/v1/sys/user/getUserList"},
		{"POST", "/api/v1/sys/user/logout"},
		{"PUT", "/api/v1/sys/user/info"},
		{"PUT", "/api/v1/sys/user/ui-config"},
		{"POST", "/api/v1/sys/user/avatar"},
		{"POST", "/api/v1/sys/user/switchAuthority"},
		{"POST", "/api/v1/sys/user/addUser"},
		{"PUT", "/api/v1/sys/user/updateUser"},
		{"DELETE", "/api/v1/sys/user/deleteUser"},
		{"POST", "/api/v1/sys/user/resetPassword"},
		{"GET", "/api/v1/sys/menu/getMenu"},
		{"POST", "/api/v1/sys/menu/getMenuList"},
		{"POST", "/api/v1/sys/menu/getMenuAuthority"},
		{"POST", "/api/v1/sys/menu/addBaseMenu"},
		{"PUT", "/api/v1/sys/menu/updateBaseMenu"},
		{"DELETE", "/api/v1/sys/menu/deleteBaseMenu"},
		{"POST", "/api/v1/sys/authority/getAuthorityList"},
		{"POST", "/api/v1/sys/authority/createAuthority"},
		{"PUT", "/api/v1/sys/authority/updateAuthority"},
		{"DELETE", "/api/v1/sys/authority/deleteAuthority"},
		{"POST", "/api/v1/sys/authority/setAuthorityMenus"},
		{"POST", "/api/v1/sys/api/getApiList"},
		{"POST", "/api/v1/sys/api/createApi"},
		{"PUT", "/api/v1/sys/api/updateApi"},
		{"DELETE", "/api/v1/sys/api/deleteApi"},
		{"POST", "/api/v1/sys/casbin/getPolicyPathByAuthorityId"},
		{"POST", "/api/v1/sys/casbin/updateCasbin"},
		{"POST", "/api/v1/sys/operationLog/getOperationLogList"},
		{"DELETE", "/api/v1/sys/operationLog/deleteOperationLogByIds"},
		{"POST", "/api/v1/sys/file/upload"},
		{"POST", "/api/v1/sys/system/getServerInfo"},
		{"POST", "/api/v1/sys/notice/createNotice"},
		{"POST", "/api/v1/sys/notice/getNoticeList"},
		{"GET", "/api/v1/sys/notice/getMyNotices"},
		{"POST", "/api/v1/sys/notice/markRead"},
		{"POST", "/api/v1/plugin/plugin/getPluginList"},
		{"POST", "/api/v1/plugin/plugin/createPlugin"},
		{"PUT", "/api/v1/plugin/plugin/updatePlugin"},
		{"POST", "/api/v1/plugin/release/getReleaseList"},
		{"POST", "/api/v1/plugin/release/getReleaseDetail"},
		{"POST", "/api/v1/plugin/release/createRelease"},
		{"PUT", "/api/v1/plugin/release/updateRelease"},
		{"POST", "/api/v1/plugin/release/transition"},
		{"POST", "/api/v1/poetry/dynasty"},
		{"PUT", "/api/v1/poetry/dynasty/:id"},
		{"DELETE", "/api/v1/poetry/dynasty/:id"},
		{"GET", "/api/v1/poetry/dynasty/list"},
		{"GET", "/api/v1/poetry/dynasty/all"},
		{"POST", "/api/v1/poetry/genre"},
		{"PUT", "/api/v1/poetry/genre/:id"},
		{"DELETE", "/api/v1/poetry/genre/:id"},
		{"GET", "/api/v1/poetry/genre/list"},
		{"GET", "/api/v1/poetry/genre/all"},
		{"POST", "/api/v1/poetry/author"},
		{"PUT", "/api/v1/poetry/author/:id"},
		{"DELETE", "/api/v1/poetry/author/:id"},
		{"GET", "/api/v1/poetry/author/list"},
		{"GET", "/api/v1/poetry/author/:id"},
		{"POST", "/api/v1/poetry/author/avatar"},
		{"POST", "/api/v1/poetry/poem"},
		{"PUT", "/api/v1/poetry/poem/:id"},
		{"DELETE", "/api/v1/poetry/poem/:id"},
		{"GET", "/api/v1/poetry/poem/list"},
		{"GET", "/api/v1/poetry/poem/:id"},
	}
	basicAccess := [][]string{
		{"GET", "/api/v1/sys/user/getSelfInfo"},
		{"POST", "/api/v1/sys/user/logout"},
		{"PUT", "/api/v1/sys/user/info"},
		{"PUT", "/api/v1/sys/user/ui-config"},
		{"POST", "/api/v1/sys/user/avatar"},
		{"GET", "/api/v1/sys/menu/getMenu"},
		{"POST", "/api/v1/sys/system/getServerInfo"},
		{"GET", "/api/v1/sys/notice/getMyNotices"},
		{"POST", "/api/v1/sys/notice/markRead"},
		{"GET", "/api/v1/poetry/dynasty/list"},
		{"GET", "/api/v1/poetry/dynasty/all"},
		{"GET", "/api/v1/poetry/genre/list"},
		{"GET", "/api/v1/poetry/genre/all"},
		{"GET", "/api/v1/poetry/author/list"},
		{"GET", "/api/v1/poetry/author/:id"},
		{"GET", "/api/v1/poetry/poem/list"},
		{"GET", "/api/v1/poetry/poem/:id"},
	}

	rolePolicies := map[uint][][]string{
		adminAuthorityID: fullAccess,
		9528:             fullAccess,
		888:              basicAccess,
		8881:             basicAccess,
		10010:            pluginOnlyPolicyWithBasic(basicAccess),
		10013:            pluginOnlyPolicyWithBasic(basicAccess),
		10014:            pluginOnlyPolicyWithBasic(basicAccess),
	}

	rules := make([]casbinRuleSeed, 0)
	for authorityID, policies := range rolePolicies {
		sub := strconv.FormatUint(uint64(authorityID), 10)
		for _, p := range policies {
			rules = append(rules, casbinRuleSeed{Ptype: "p", V0: sub, V1: p[1], V2: p[0], V3: "", V4: "", V5: ""})
		}
	}
	if len(rules) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rules).Error
}

func ensureAdminUser(tx *gorm.DB, opts seedAdminOptions) error {
	var existing model.SysUser
	err := tx.Where("username = ?", opts.Username).First(&existing).Error
	if err == nil {
		if existing.AuthorityID == 0 {
			if err := tx.Model(&existing).Update("authority_id", opts.AuthorityID).Error; err != nil {
				return err
			}
		}
		userAuth := model.SysUserAuthority{UserId: existing.ID, AuthorityId: opts.AuthorityID}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&userAuth).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPwd, err := utils.BcryptHash(opts.Password)
	if err != nil {
		return err
	}

	admin := model.SysUser{
		UUID:        uuid.New(),
		Username:    opts.Username,
		Password:    hashedPwd,
		NickName:    "Administrator",
		Avatar:      model.DefaultUserAvatar,
		Status:      model.UserActive,
		AuthorityID: opts.AuthorityID,
	}
	if err := tx.Create(&admin).Error; err != nil {
		return err
	}

	userAuth := model.SysUserAuthority{UserId: admin.ID, AuthorityId: opts.AuthorityID}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&userAuth).Error
}

func apiSign(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func pluginOnlyAccess() []string {
	return []string{
		apiSign("POST", "/api/v1/plugin/plugin/getPluginList"),
		apiSign("POST", "/api/v1/plugin/plugin/getProjectDetail"),
		apiSign("POST", "/api/v1/plugin/plugin/createPlugin"),
		apiSign("PUT", "/api/v1/plugin/plugin/updatePlugin"),
		apiSign("POST", "/api/v1/plugin/release/getReleaseList"),
		apiSign("POST", "/api/v1/plugin/release/getReleaseDetail"),
		apiSign("POST", "/api/v1/plugin/release/createRelease"),
		apiSign("PUT", "/api/v1/plugin/release/updateRelease"),
		apiSign("POST", "/api/v1/plugin/release/transition"),
	}
}

func pluginOnlyPolicyWithBasic(basic [][]string) [][]string {
	policies := make([][]string, 0, len(basic)+8)
	policies = append(policies, basic...)
	policies = append(policies,
		[]string{"POST", "/api/v1/plugin/plugin/getPluginList"},
		[]string{"POST", "/api/v1/plugin/plugin/getProjectDetail"},
		[]string{"POST", "/api/v1/plugin/plugin/createPlugin"},
		[]string{"PUT", "/api/v1/plugin/plugin/updatePlugin"},
		[]string{"POST", "/api/v1/plugin/release/getReleaseList"},
		[]string{"POST", "/api/v1/plugin/release/getReleaseDetail"},
		[]string{"POST", "/api/v1/plugin/release/createRelease"},
		[]string{"PUT", "/api/v1/plugin/release/updateRelease"},
		[]string{"POST", "/api/v1/plugin/release/transition"},
	)
	return policies
}

func ensurePluginBaseData(tx *gorm.DB) error {
	userIDs, err := ensurePluginDemoUsers(tx)
	if err != nil {
		return err
	}
	pluginIDs, err := ensurePluginCatalog(tx, userIDs["requester"])
	if err != nil {
		return err
	}
	return ensurePluginReleaseSeeds(tx, pluginIDs, userIDs)
}

func ensurePluginDemoUsers(tx *gorm.DB) (map[string]uint, error) {
	seeds := []struct {
		Key         string
		Username    string
		Nickname    string
		AuthorityID uint
	}{
		{Key: "requester", Username: "plugin.requester", Nickname: "插件提单人", AuthorityID: 10010},
		{Key: "reviewer", Username: "plugin.reviewer", Nickname: "插件审核人", AuthorityID: 10013},
		{Key: "publisher", Username: "plugin.publisher", Nickname: "插件发布管理员", AuthorityID: 10014},
	}

	result := make(map[string]uint, len(seeds))
	for _, seed := range seeds {
		userID, err := ensureSeedUser(tx, seed.Username, seed.Nickname, seed.AuthorityID, "Plugin@123456")
		if err != nil {
			return nil, err
		}
		result[seed.Key] = userID
	}
	return result, nil
}

func ensureSeedUser(tx *gorm.DB, username, nickname string, authorityID uint, password string) (uint, error) {
	var user model.SysUser
	err := tx.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hashedPwd, hashErr := utils.BcryptHash(password)
		if hashErr != nil {
			return 0, hashErr
		}
		user = model.SysUser{
			UUID:        uuid.New(),
			Username:    username,
			Password:    hashedPwd,
			NickName:    nickname,
			Avatar:      model.DefaultUserAvatar,
			Status:      model.UserActive,
			AuthorityID: authorityID,
		}
		if err := tx.Create(&user).Error; err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	} else {
		updates := map[string]interface{}{
			"nick_name":    nickname,
			"status":       model.UserActive,
			"authority_id": authorityID,
		}
		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return 0, err
		}
	}

	relation := model.SysUserAuthority{UserId: user.ID, AuthorityId: authorityID}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&relation).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func ensurePluginCatalog(tx *gorm.DB, creatorID uint) (map[string]uint, error) {
	seeds := []pluginModel.Plugin{
		{
			Code:          "smart-audit",
			RepositoryURL: "https://git.example.com/plugins/smart-audit.git",
			NameZh:        "智能审计助手",
			NameEn:        "Smart Audit Assistant",
			DescriptionZh: "用于对业务流程和插件行为做规则扫描与审计的插件。",
			DescriptionEn: "A plugin for rule-based scanning and auditing of business flows and plugin behavior.",
			CapabilityZh:  "规则扫描、风险识别、审计报告导出",
			CapabilityEn:  "Rule scanning, risk detection, audit report export",
			Owner:         "平台治理组",
			CreatedBy:     creatorID,
			CurrentStatus: pluginModel.PluginStatusActive,
		},
		{
			Code:          "image-optimizer",
			RepositoryURL: "https://git.example.com/plugins/image-optimizer.git",
			NameZh:        "图片优化引擎",
			NameEn:        "Image Optimizer Engine",
			DescriptionZh: "对上传图片做格式转换、压缩与清晰度增强。",
			DescriptionEn: "A plugin for image conversion, compression, and clarity enhancement.",
			CapabilityZh:  "WebP/AVIF 转码、批量压缩、质量分析",
			CapabilityEn:  "WebP/AVIF transcoding, batch compression, quality analysis",
			Owner:         "多媒体平台组",
			CreatedBy:     creatorID,
			CurrentStatus: pluginModel.PluginStatusActive,
		},
		{
			Code:          "device-bridge",
			RepositoryURL: "https://git.example.com/plugins/device-bridge.git",
			NameZh:        "设备桥接服务",
			NameEn:        "Device Bridge Service",
			DescriptionZh: "为边缘设备接入提供标准连接协议与远程指令能力。",
			DescriptionEn: "A standard connectivity and remote command plugin for edge device onboarding.",
			CapabilityZh:  "协议适配、远程指令、设备元数据同步",
			CapabilityEn:  "Protocol adaptation, remote commands, device metadata sync",
			Owner:         "IoT 接入组",
			CreatedBy:     creatorID,
			CurrentStatus: pluginModel.PluginStatusPlanning,
		},
		{
			Code:          "edge-runtime",
			RepositoryURL: "https://git.example.com/plugins/edge-runtime.git",
			NameZh:        "边缘运行时",
			NameEn:        "Edge Runtime",
			DescriptionZh: "提供多架构插件包的边缘执行环境与基础资源管理。",
			DescriptionEn: "A multi-architecture edge runtime for plugin execution and resource management.",
			CapabilityZh:  "x86/ARM 运行时、资源限额、日志采集",
			CapabilityEn:  "x86/ARM runtime, resource quotas, log collection",
			Owner:         "边缘计算组",
			CreatedBy:     creatorID,
			CurrentStatus: pluginModel.PluginStatusPlanning,
		},
		{
			Code:          "legacy-sync",
			RepositoryURL: "https://git.example.com/plugins/legacy-sync.git",
			NameZh:        "历史系统同步器",
			NameEn:        "Legacy Sync Connector",
			DescriptionZh: "同步历史系统中的订单与客户数据，现已进入淘汰流程。",
			DescriptionEn: "A connector for syncing orders and customers from a legacy system, now under retirement.",
			CapabilityZh:  "订单同步、客户同步、补偿任务",
			CapabilityEn:  "Order sync, customer sync, compensation jobs",
			Owner:         "集成平台组",
			CreatedBy:     creatorID,
			CurrentStatus: pluginModel.PluginStatusOfflined,
		},
		{
			Code:          "content-moderation",
			RepositoryURL: "https://git.example.com/plugins/content-moderation.git",
			NameZh:        "内容审核网关",
			NameEn:        "Content Moderation Gateway",
			DescriptionZh: "对文本与图片内容做统一审核编排，当前处于发布准备阶段。",
			DescriptionEn: "A unified moderation gateway for text and image review, currently in release preparation.",
			CapabilityZh:  "文本审核、图片审核、策略路由",
			CapabilityEn:  "Text moderation, image moderation, policy routing",
			Owner:         "内容安全组",
			CreatedBy:     creatorID,
			CurrentStatus: pluginModel.PluginStatusPlanning,
		},
	}

	result := make(map[string]uint, len(seeds))
	for _, seed := range seeds {
		var item pluginModel.Plugin
		err := tx.Where("code = ?", seed.Code).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = seed
			if err := tx.Create(&item).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			updates := map[string]interface{}{
				"repository_url": seed.RepositoryURL,
				"name_zh":        seed.NameZh,
				"name_en":        seed.NameEn,
				"description_zh": seed.DescriptionZh,
				"description_en": seed.DescriptionEn,
				"capability_zh":  seed.CapabilityZh,
				"capability_en":  seed.CapabilityEn,
				"owner":          seed.Owner,
				"created_by":     creatorID,
			}
			if err := tx.Model(&item).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		result[seed.Code] = item.ID
	}
	return result, nil
}

func ensurePluginReleaseSeeds(tx *gorm.DB, pluginIDs map[string]uint, userIDs map[string]uint) error {
	releasedChecklist, err := buildSampleChecklistJSON()
	if err != nil {
		return err
	}

	smartAuditRelease, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:             pluginIDs["smart-audit"],
		RequestType:          pluginModel.PluginReleaseTypeInitial,
		Status:               pluginModel.PluginReleaseStatusReleased,
		Version:              "1.0.0",
		VersionConstraint:    ">= web-cms 2.0.0",
		Publisher:            "plugin.publisher",
		ReviewerID:           uintPtr(userIDs["reviewer"]),
		PublisherID:          uintPtr(userIDs["publisher"]),
		Checklist:            releasedChecklist,
		PerformanceSummaryZh: "核心规则扫描任务稳定，资源开销符合发布门槛。",
		PerformanceSummaryEn: "Core scanning tasks are stable and resource consumption meets the release baseline.",
		TestReportURL:        "https://minio.local/gowebframe/plugin-demo/smart-audit/test-report-1.0.0.pdf",
		PackageX86URL:        "https://minio.local/gowebframe/plugin-demo/smart-audit/smart-audit-1.0.0-x86.tar.gz",
		PackageArmURL:        "https://minio.local/gowebframe/plugin-demo/smart-audit/smart-audit-1.0.0-arm.tar.gz",
		ChangelogZh:          "首发版本，提供规则扫描、风险识别与审计报告导出能力。",
		ChangelogEn:          "Initial release with rule scanning, risk detection, and audit report export.",
		CreatedBy:            userIDs["requester"],
		SubmittedAt:          timePtr(time.Now().Add(-96 * time.Hour)),
		ApprovedAt:           timePtr(time.Now().Add(-95 * time.Hour)),
		ReleasedAt:           timePtr(time.Now().Add(-94 * time.Hour)),
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusReleasePreparing, OperatorID: userIDs["requester"], Comment: "首发工单已创建"},
		{Action: "submit_review", FromStatus: pluginModel.PluginReleaseStatusReleasePreparing, ToStatus: pluginModel.PluginReleaseStatusPendingReview, OperatorID: userIDs["requester"], Comment: "发布资料已齐备，提交审核"},
		{Action: "approve", FromStatus: pluginModel.PluginReleaseStatusPendingReview, ToStatus: pluginModel.PluginReleaseStatusApproved, OperatorID: userIDs["reviewer"], Comment: "审核通过，可进入发布"},
		{Action: "release", FromStatus: pluginModel.PluginReleaseStatusApproved, ToStatus: pluginModel.PluginReleaseStatusReleased, OperatorID: userIDs["publisher"], Comment: "已执行正式发布"},
	})
	if err != nil {
		return err
	}

	if _, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:             pluginIDs["smart-audit"],
		RequestType:          pluginModel.PluginReleaseTypeMaintenance,
		Status:               pluginModel.PluginReleaseStatusRejected,
		SourceReleaseID:      uintPtr(smartAuditRelease.ID),
		Version:              "1.0.1",
		VersionConstraint:    ">= web-cms 2.0.0",
		Publisher:            "plugin.publisher",
		ReviewerID:           uintPtr(userIDs["reviewer"]),
		PublisherID:          uintPtr(userIDs["publisher"]),
		Checklist:            releasedChecklist,
		PerformanceSummaryZh: "修复规则误报问题，但边界案例仍需补充验证。",
		PerformanceSummaryEn: "Fixes false positives, but edge-case validation still needs more coverage.",
		TestReportURL:        "https://minio.local/gowebframe/plugin-demo/smart-audit/test-report-1.0.1.pdf",
		PackageX86URL:        "https://minio.local/gowebframe/plugin-demo/smart-audit/smart-audit-1.0.1-x86.tar.gz",
		PackageArmURL:        "https://minio.local/gowebframe/plugin-demo/smart-audit/smart-audit-1.0.1-arm.tar.gz",
		ChangelogZh:          "修复规则误报并补充告警分级能力。",
		ChangelogEn:          "Fixes false positives and adds alert severity levels.",
		ReviewComment:        "需要补充中文和英文性能对比说明后再提交。",
		CreatedBy:            userIDs["requester"],
		SubmittedAt:          timePtr(time.Now().Add(-48 * time.Hour)),
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusReleasePreparing, OperatorID: userIDs["requester"], Comment: "维护工单已创建"},
		{Action: "submit_review", FromStatus: pluginModel.PluginReleaseStatusReleasePreparing, ToStatus: pluginModel.PluginReleaseStatusPendingReview, OperatorID: userIDs["requester"], Comment: "提交维护版本审核"},
		{Action: "reject", FromStatus: pluginModel.PluginReleaseStatusPendingReview, ToStatus: pluginModel.PluginReleaseStatusRejected, OperatorID: userIDs["reviewer"], Comment: "需要补充双语性能说明"},
	}); err != nil {
		return err
	}

	if _, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:             pluginIDs["image-optimizer"],
		RequestType:          pluginModel.PluginReleaseTypeInitial,
		Status:               pluginModel.PluginReleaseStatusReleased,
		Version:              "1.4.2",
		VersionConstraint:    ">= web-cms 1.9.0",
		Publisher:            "plugin.publisher",
		ReviewerID:           uintPtr(userIDs["reviewer"]),
		PublisherID:          uintPtr(userIDs["publisher"]),
		Checklist:            releasedChecklist,
		PerformanceSummaryZh: "压缩链路耗时稳定，图像质量评分达到上线要求。",
		PerformanceSummaryEn: "Compression latency is stable and quality scores meet production requirements.",
		TestReportURL:        "https://minio.local/gowebframe/plugin-demo/image-optimizer/test-report-1.4.2.pdf",
		PackageX86URL:        "https://minio.local/gowebframe/plugin-demo/image-optimizer/image-optimizer-1.4.2-x86.tar.gz",
		PackageArmURL:        "https://minio.local/gowebframe/plugin-demo/image-optimizer/image-optimizer-1.4.2-arm.tar.gz",
		ChangelogZh:          "增强 WebP/AVIF 转码效果并优化批量压缩吞吐。",
		ChangelogEn:          "Improves WebP/AVIF transcoding and batch compression throughput.",
		CreatedBy:            userIDs["requester"],
		SubmittedAt:          timePtr(time.Now().Add(-168 * time.Hour)),
		ApprovedAt:           timePtr(time.Now().Add(-167 * time.Hour)),
		ReleasedAt:           timePtr(time.Now().Add(-166 * time.Hour)),
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusReleasePreparing, OperatorID: userIDs["requester"], Comment: "图片优化插件首发工单"},
		{Action: "submit_review", FromStatus: pluginModel.PluginReleaseStatusReleasePreparing, ToStatus: pluginModel.PluginReleaseStatusPendingReview, OperatorID: userIDs["requester"], Comment: "已上传多架构包和测试报告"},
		{Action: "approve", FromStatus: pluginModel.PluginReleaseStatusPendingReview, ToStatus: pluginModel.PluginReleaseStatusApproved, OperatorID: userIDs["reviewer"], Comment: "审核通过"},
		{Action: "release", FromStatus: pluginModel.PluginReleaseStatusApproved, ToStatus: pluginModel.PluginReleaseStatusReleased, OperatorID: userIDs["publisher"], Comment: "正式发布完成"},
	}); err != nil {
		return err
	}

	if _, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:             pluginIDs["device-bridge"],
		RequestType:          pluginModel.PluginReleaseTypeInitial,
		Status:               pluginModel.PluginReleaseStatusPendingReview,
		Version:              "2.3.0",
		VersionConstraint:    ">= web-cms 2.1.0",
		Publisher:            "plugin.publisher",
		ReviewerID:           uintPtr(userIDs["reviewer"]),
		PublisherID:          uintPtr(userIDs["publisher"]),
		Checklist:            releasedChecklist,
		PerformanceSummaryZh: "连接稳定性通过基线，待审核后进入正式发布。",
		PerformanceSummaryEn: "Connection stability passed the baseline and is awaiting final review.",
		TestReportURL:        "https://minio.local/gowebframe/plugin-demo/device-bridge/test-report-2.3.0.pdf",
		PackageX86URL:        "https://minio.local/gowebframe/plugin-demo/device-bridge/device-bridge-2.3.0-x86.tar.gz",
		PackageArmURL:        "https://minio.local/gowebframe/plugin-demo/device-bridge/device-bridge-2.3.0-arm.tar.gz",
		ChangelogZh:          "新增标准设备接入协议适配与远程指令能力。",
		ChangelogEn:          "Adds standard device protocol adaptation and remote command support.",
		CreatedBy:            userIDs["requester"],
		SubmittedAt:          timePtr(time.Now().Add(-12 * time.Hour)),
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusReleasePreparing, OperatorID: userIDs["requester"], Comment: "设备桥接插件工单"},
		{Action: "submit_review", FromStatus: pluginModel.PluginReleaseStatusReleasePreparing, ToStatus: pluginModel.PluginReleaseStatusPendingReview, OperatorID: userIDs["requester"], Comment: "待审核人处理"},
	}); err != nil {
		return err
	}

	if _, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:             pluginIDs["edge-runtime"],
		RequestType:          pluginModel.PluginReleaseTypeInitial,
		Status:               pluginModel.PluginReleaseStatusApproved,
		Version:              "0.9.0",
		VersionConstraint:    ">= web-cms 2.1.0",
		Publisher:            "plugin.publisher",
		ReviewerID:           uintPtr(userIDs["reviewer"]),
		PublisherID:          uintPtr(userIDs["publisher"]),
		Checklist:            releasedChecklist,
		PerformanceSummaryZh: "运行时稳定，等待发布管理员执行正式发布。",
		PerformanceSummaryEn: "Runtime is stable and waiting for the publisher to execute release.",
		TestReportURL:        "https://minio.local/gowebframe/plugin-demo/edge-runtime/test-report-0.9.0.pdf",
		PackageX86URL:        "https://minio.local/gowebframe/plugin-demo/edge-runtime/edge-runtime-0.9.0-x86.tar.gz",
		PackageArmURL:        "https://minio.local/gowebframe/plugin-demo/edge-runtime/edge-runtime-0.9.0-arm.tar.gz",
		ChangelogZh:          "提供边缘运行时基础能力与资源治理。",
		ChangelogEn:          "Provides baseline edge runtime capabilities and resource governance.",
		CreatedBy:            userIDs["requester"],
		SubmittedAt:          timePtr(time.Now().Add(-10 * time.Hour)),
		ApprovedAt:           timePtr(time.Now().Add(-9 * time.Hour)),
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusReleasePreparing, OperatorID: userIDs["requester"], Comment: "边缘运行时首发工单"},
		{Action: "submit_review", FromStatus: pluginModel.PluginReleaseStatusReleasePreparing, ToStatus: pluginModel.PluginReleaseStatusPendingReview, OperatorID: userIDs["requester"], Comment: "已提交审核"},
		{Action: "approve", FromStatus: pluginModel.PluginReleaseStatusPendingReview, ToStatus: pluginModel.PluginReleaseStatusApproved, OperatorID: userIDs["reviewer"], Comment: "等待发布执行"},
	}); err != nil {
		return err
	}

	legacyRelease, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:             pluginIDs["legacy-sync"],
		RequestType:          pluginModel.PluginReleaseTypeInitial,
		Status:               pluginModel.PluginReleaseStatusReleased,
		Version:              "3.2.1",
		VersionConstraint:    ">= web-cms 1.8.0",
		Publisher:            "plugin.publisher",
		ReviewerID:           uintPtr(userIDs["reviewer"]),
		PublisherID:          uintPtr(userIDs["publisher"]),
		Checklist:            releasedChecklist,
		PerformanceSummaryZh: "该版本曾经稳定发布，现因重大风险进入下架流程。",
		PerformanceSummaryEn: "This version was previously stable, but is now being retired due to critical risks.",
		TestReportURL:        "https://minio.local/gowebframe/plugin-demo/legacy-sync/test-report-3.2.1.pdf",
		PackageX86URL:        "https://minio.local/gowebframe/plugin-demo/legacy-sync/legacy-sync-3.2.1-x86.tar.gz",
		PackageArmURL:        "https://minio.local/gowebframe/plugin-demo/legacy-sync/legacy-sync-3.2.1-arm.tar.gz",
		ChangelogZh:          "历史版本，用于演示下架流程。",
		ChangelogEn:          "Legacy version kept for offlining workflow demonstration.",
		IsOfflined:           true,
		CreatedBy:            userIDs["requester"],
		SubmittedAt:          timePtr(time.Now().Add(-240 * time.Hour)),
		ApprovedAt:           timePtr(time.Now().Add(-239 * time.Hour)),
		ReleasedAt:           timePtr(time.Now().Add(-238 * time.Hour)),
		OfflinedAt:           timePtr(time.Now().Add(-24 * time.Hour)),
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusReleasePreparing, OperatorID: userIDs["requester"], Comment: "历史同步器首发工单"},
		{Action: "submit_review", FromStatus: pluginModel.PluginReleaseStatusReleasePreparing, ToStatus: pluginModel.PluginReleaseStatusPendingReview, OperatorID: userIDs["requester"], Comment: "已提交审核"},
		{Action: "approve", FromStatus: pluginModel.PluginReleaseStatusPendingReview, ToStatus: pluginModel.PluginReleaseStatusApproved, OperatorID: userIDs["reviewer"], Comment: "审核通过"},
		{Action: "release", FromStatus: pluginModel.PluginReleaseStatusApproved, ToStatus: pluginModel.PluginReleaseStatusReleased, OperatorID: userIDs["publisher"], Comment: "曾经正式发布"},
	})
	if err != nil {
		return err
	}

	if _, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:        pluginIDs["legacy-sync"],
		RequestType:     pluginModel.PluginReleaseTypeOffline,
		Status:          pluginModel.PluginReleaseStatusOfflined,
		TargetReleaseID: uintPtr(legacyRelease.ID),
		ReviewerID:      uintPtr(userIDs["reviewer"]),
		PublisherID:     uintPtr(userIDs["publisher"]),
		OfflineReasonZh: "发现历史同步任务可能造成脏数据回写，需紧急下架。",
		OfflineReasonEn: "A critical risk was found where sync jobs may write back dirty data, so the version must be offlined.",
		ReviewComment:   "风险确认，批准下架",
		CreatedBy:       userIDs["requester"],
		SubmittedAt:     timePtr(time.Now().Add(-26 * time.Hour)),
		ApprovedAt:      timePtr(time.Now().Add(-25 * time.Hour)),
		OfflinedAt:      timePtr(time.Now().Add(-24 * time.Hour)),
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusDraft, OperatorID: userIDs["requester"], Comment: "发起下架工单"},
		{Action: "submit_review", FromStatus: pluginModel.PluginReleaseStatusDraft, ToStatus: pluginModel.PluginReleaseStatusPendingReview, OperatorID: userIDs["requester"], Comment: "提交下架审核"},
		{Action: "approve", FromStatus: pluginModel.PluginReleaseStatusPendingReview, ToStatus: pluginModel.PluginReleaseStatusApproved, OperatorID: userIDs["reviewer"], Comment: "审核通过"},
		{Action: "release", FromStatus: pluginModel.PluginReleaseStatusApproved, ToStatus: pluginModel.PluginReleaseStatusOfflined, OperatorID: userIDs["publisher"], Comment: "已执行下架"},
	}); err != nil {
		return err
	}

	if _, err := ensureSeedRelease(tx, &pluginModel.PluginRelease{
		PluginID:             pluginIDs["content-moderation"],
		RequestType:          pluginModel.PluginReleaseTypeInitial,
		Status:               pluginModel.PluginReleaseStatusReleasePreparing,
		Version:              "0.5.0",
		VersionConstraint:    ">= web-cms 2.1.0",
		Publisher:            "plugin.publisher",
		ReviewerID:           uintPtr(userIDs["reviewer"]),
		PublisherID:          uintPtr(userIDs["publisher"]),
		Checklist:            releasedChecklist,
		PerformanceSummaryZh: "资料准备中，等待提单人补齐变更说明后提交审核。",
		PerformanceSummaryEn: "Release materials are being prepared and waiting for the requester to complete changelog details.",
		TestReportURL:        "https://minio.local/gowebframe/plugin-demo/content-moderation/test-report-0.5.0.pdf",
		PackageX86URL:        "https://minio.local/gowebframe/plugin-demo/content-moderation/content-moderation-0.5.0-x86.tar.gz",
		PackageArmURL:        "https://minio.local/gowebframe/plugin-demo/content-moderation/content-moderation-0.5.0-arm.tar.gz",
		CreatedBy:            userIDs["requester"],
	}, []seedReleaseEvent{
		{Action: "create", ToStatus: pluginModel.PluginReleaseStatusReleasePreparing, OperatorID: userIDs["requester"], Comment: "首发版本资料准备中"},
	}); err != nil {
		return err
	}

	if err := tx.Model(&pluginModel.Plugin{}).Where("id = ?", pluginIDs["smart-audit"]).Updates(map[string]interface{}{
		"current_status":   pluginModel.PluginStatusActive,
		"latest_version":   "1.0.0",
		"last_released_at": time.Now().Add(-94 * time.Hour),
	}).Error; err != nil {
		return err
	}
	if err := tx.Model(&pluginModel.Plugin{}).Where("id = ?", pluginIDs["image-optimizer"]).Updates(map[string]interface{}{
		"current_status":   pluginModel.PluginStatusActive,
		"latest_version":   "1.4.2",
		"last_released_at": time.Now().Add(-166 * time.Hour),
	}).Error; err != nil {
		return err
	}
	if err := tx.Model(&pluginModel.Plugin{}).Where("id = ?", pluginIDs["device-bridge"]).Updates(map[string]interface{}{
		"current_status": pluginModel.PluginStatusPlanning,
	}).Error; err != nil {
		return err
	}
	if err := tx.Model(&pluginModel.Plugin{}).Where("id = ?", pluginIDs["edge-runtime"]).Updates(map[string]interface{}{
		"current_status": pluginModel.PluginStatusPlanning,
	}).Error; err != nil {
		return err
	}
	if err := tx.Model(&pluginModel.Plugin{}).Where("id = ?", pluginIDs["legacy-sync"]).Updates(map[string]interface{}{
		"current_status":   pluginModel.PluginStatusOfflined,
		"latest_version":   "3.2.1",
		"last_released_at": time.Now().Add(-238 * time.Hour),
	}).Error; err != nil {
		return err
	}
	if err := tx.Model(&pluginModel.Plugin{}).Where("id = ?", pluginIDs["content-moderation"]).Updates(map[string]interface{}{
		"current_status": pluginModel.PluginStatusPlanning,
	}).Error; err != nil {
		return err
	}

	return nil
}

type seedReleaseEvent struct {
	Action     string
	FromStatus pluginModel.PluginReleaseStatus
	ToStatus   pluginModel.PluginReleaseStatus
	OperatorID uint
	Comment    string
}

func ensureSeedRelease(tx *gorm.DB, seed *pluginModel.PluginRelease, events []seedReleaseEvent) (*pluginModel.PluginRelease, error) {
	var item pluginModel.PluginRelease
	query := tx.Where("plugin_id = ? AND request_type = ?", seed.PluginID, seed.RequestType)
	if seed.RequestType == pluginModel.PluginReleaseTypeOffline && seed.TargetReleaseID != nil {
		query = query.Where("target_release_id = ?", *seed.TargetReleaseID)
	} else {
		query = query.Where("version = ?", seed.Version)
	}

	err := query.First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item = *seed
		if err := tx.Create(&item).Error; err != nil {
			return nil, err
		}
		for _, event := range events {
			if err := appendSeedReleaseEvent(tx, item.ID, event); err != nil {
				return nil, err
			}
		}
		return &item, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func appendSeedReleaseEvent(tx *gorm.DB, releaseID uint, event seedReleaseEvent) error {
	record := pluginModel.PluginReleaseEvent{
		ReleaseID:    releaseID,
		FromStatus:   event.FromStatus,
		ToStatus:     event.ToStatus,
		Action:       event.Action,
		OperatorID:   event.OperatorID,
		Comment:      event.Comment,
		SnapshotJSON: datatypes.JSON([]byte("{}")),
	}
	return tx.Create(&record).Error
}

func buildSampleChecklistJSON() (datatypes.JSON, error) {
	raw, err := json.Marshal([]pluginModel.PluginChecklistItem{
		{
			TitleZh: "双语发布资料齐全",
			TitleEn: "Bilingual release materials completed",
			Passed:  true,
			NoteZh:  "名称、描述、能力、变更说明均已补齐",
			NoteEn:  "Name, description, capabilities, and changelog are complete",
		},
		{
			TitleZh: "多架构安装包验证",
			TitleEn: "Multi-architecture package validation",
			Passed:  true,
			NoteZh:  "x86 和 ARM 包均可正常安装",
			NoteEn:  "Both x86 and ARM packages can be installed successfully",
		},
	})
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(raw), nil
}

func uintPtr(v uint) *uint {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}

func ensurePoetryBaseData(tx *gorm.DB) error {
	dynasties, err := ensureDynasties(tx)
	if err != nil {
		return err
	}
	genres, err := ensureGenres(tx)
	if err != nil {
		return err
	}
	tags, err := ensureTags(tx)
	if err != nil {
		return err
	}
	authors, err := ensureAuthors(tx, dynasties)
	if err != nil {
		return err
	}
	works, err := ensureWorks(tx, authors, genres)
	if err != nil {
		return err
	}
	return ensureWorkTags(tx, works, tags)
}

func ensureDynasties(tx *gorm.DB) (map[string]uint, error) {
	seeds := []poetryModel.MetaDynasty{
		{Name: "先秦", SortOrder: 1},
		{Name: "汉", SortOrder: 2},
		{Name: "唐", SortOrder: 3},
		{Name: "宋", SortOrder: 4},
		{Name: "元", SortOrder: 5},
		{Name: "明", SortOrder: 6},
		{Name: "清", SortOrder: 7},
	}
	result := make(map[string]uint, len(seeds))
	for _, seed := range seeds {
		var item poetryModel.MetaDynasty
		err := tx.Where("name = ?", seed.Name).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = seed
			if err := tx.Create(&item).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else if item.SortOrder != seed.SortOrder {
			if err := tx.Model(&item).Update("sort_order", seed.SortOrder).Error; err != nil {
				return nil, err
			}
		}
		result[seed.Name] = item.ID
	}
	return result, nil
}

func ensureGenres(tx *gorm.DB) (map[string]uint, error) {
	seeds := []poetryModel.MetaGenre{
		{Name: "诗", SortOrder: 1},
		{Name: "词", SortOrder: 2},
		{Name: "曲", SortOrder: 3},
		{Name: "赋", SortOrder: 4},
	}
	result := make(map[string]uint, len(seeds))
	for _, seed := range seeds {
		var item poetryModel.MetaGenre
		err := tx.Where("name = ?", seed.Name).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = seed
			if err := tx.Create(&item).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else if item.SortOrder != seed.SortOrder {
			if err := tx.Model(&item).Update("sort_order", seed.SortOrder).Error; err != nil {
				return nil, err
			}
		}
		result[seed.Name] = item.ID
	}
	return result, nil
}

func ensureTags(tx *gorm.DB) (map[string]uint, error) {
	seeds := []poetryModel.MetaTag{
		{Name: "山水", Category: "题材", SortOrder: 1},
		{Name: "思乡", Category: "情感", SortOrder: 2},
		{Name: "边塞", Category: "题材", SortOrder: 3},
		{Name: "婉约", Category: "风格", SortOrder: 4},
		{Name: "豪放", Category: "风格", SortOrder: 5},
		{Name: "咏史", Category: "题材", SortOrder: 6},
	}
	result := make(map[string]uint, len(seeds))
	for _, seed := range seeds {
		var item poetryModel.MetaTag
		err := tx.Where("name = ?", seed.Name).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = seed
			if err := tx.Create(&item).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			updates := map[string]interface{}{
				"category":   seed.Category,
				"sort_order": seed.SortOrder,
			}
			if err := tx.Model(&item).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		result[seed.Name] = item.ID
	}
	return result, nil
}

func ensureAuthors(tx *gorm.DB, dynastyIDs map[string]uint) (map[string]uint, error) {
	type authorSeed struct {
		Name      string
		Dynasty   string
		Intro     string
		LifeStory string
	}
	seeds := []authorSeed{
		{Name: "李白", Dynasty: "唐", Intro: "字太白，号青莲居士。", LifeStory: "盛唐著名诗人，诗风飘逸豪放。"},
		{Name: "杜甫", Dynasty: "唐", Intro: "字子美，自号少陵野老。", LifeStory: "现实主义诗人，关怀家国民生。"},
		{Name: "苏轼", Dynasty: "宋", Intro: "字子瞻，号东坡居士。", LifeStory: "文学家、书画家，诗词文俱佳。"},
		{Name: "李清照", Dynasty: "宋", Intro: "号易安居士。", LifeStory: "婉约词派代表人物。"},
	}
	result := make(map[string]uint, len(seeds))
	for _, seed := range seeds {
		dynastyID, ok := dynastyIDs[seed.Dynasty]
		if !ok {
			return nil, errors.New("dynasty not found for author seed: " + seed.Dynasty)
		}
		var item poetryModel.PoemAuthor
		err := tx.Where("name = ? AND dynasty_id = ?", seed.Name, dynastyID).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = poetryModel.PoemAuthor{
				Name:      seed.Name,
				DynastyID: dynastyID,
				Intro:     seed.Intro,
				LifeStory: seed.LifeStory,
			}
			if err := tx.Create(&item).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			updates := map[string]interface{}{
				"intro":      seed.Intro,
				"life_story": seed.LifeStory,
			}
			if err := tx.Model(&item).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		result[seed.Name] = item.ID
	}
	return result, nil
}

func ensureWorks(tx *gorm.DB, authorIDs map[string]uint, genreIDs map[string]uint) (map[string]uint, error) {
	type workSeed struct {
		Title        string
		Author       string
		Genre        string
		Content      string
		Translation  string
		Annotation   string
		Appreciation string
	}
	seeds := []workSeed{
		{
			Title:       "静夜思",
			Author:      "李白",
			Genre:       "诗",
			Content:     "床前明月光，疑是地上霜。举头望明月，低头思故乡。",
			Translation: "明亮的月光洒在床前，让人疑心是地上结了霜。",
		},
		{
			Title:       "春望",
			Author:      "杜甫",
			Genre:       "诗",
			Content:     "国破山河在，城春草木深。感时花溅泪，恨别鸟惊心。",
			Translation: "山河依旧而国都破败，春天城中草木繁茂。",
		},
		{
			Title:        "念奴娇·赤壁怀古",
			Author:       "苏轼",
			Genre:        "词",
			Content:      "大江东去，浪淘尽，千古风流人物。",
			Annotation:   "赤壁：今湖北黄冈一带。",
			Appreciation: "借古抒怀，气势雄浑。",
		},
		{
			Title:        "如梦令·常记溪亭日暮",
			Author:       "李清照",
			Genre:        "词",
			Content:      "常记溪亭日暮，沉醉不知归路。",
			Appreciation: "写少女时游宴情景，清新自然。",
		},
	}
	result := make(map[string]uint, len(seeds))
	for _, seed := range seeds {
		authorID, ok := authorIDs[seed.Author]
		if !ok {
			return nil, errors.New("author not found for work seed: " + seed.Author)
		}
		genreID, ok := genreIDs[seed.Genre]
		if !ok {
			return nil, errors.New("genre not found for work seed: " + seed.Genre)
		}

		var item poetryModel.PoemWork
		err := tx.Where("title = ? AND author_id = ?", seed.Title, authorID).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = poetryModel.PoemWork{
				Title:        seed.Title,
				AuthorID:     authorID,
				GenreID:      genreID,
				Content:      seed.Content,
				Translation:  seed.Translation,
				Annotation:   seed.Annotation,
				Appreciation: seed.Appreciation,
				ViewCount:    0,
			}
			if err := tx.Create(&item).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			updates := map[string]interface{}{
				"genre_id":     genreID,
				"content":      seed.Content,
				"translation":  seed.Translation,
				"annotation":   seed.Annotation,
				"appreciation": seed.Appreciation,
			}
			if err := tx.Model(&item).Updates(updates).Error; err != nil {
				return nil, err
			}
		}
		result[seed.Title] = item.ID
	}
	return result, nil
}

func ensureWorkTags(tx *gorm.DB, workIDs map[string]uint, tagIDs map[string]uint) error {
	mapping := map[string][]string{
		"静夜思":        {"思乡"},
		"春望":         {"咏史"},
		"念奴娇·赤壁怀古":   {"豪放", "咏史"},
		"如梦令·常记溪亭日暮": {"婉约"},
	}
	relations := make([]poetryModel.PoemTagRel, 0)
	for title, tags := range mapping {
		workID, ok := workIDs[title]
		if !ok {
			continue
		}
		for _, tagName := range tags {
			tagID, ok := tagIDs[tagName]
			if !ok {
				continue
			}
			relations = append(relations, poetryModel.PoemTagRel{
				WorkID: workID,
				TagID:  tagID,
			})
		}
	}
	if len(relations) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&relations).Error
}
