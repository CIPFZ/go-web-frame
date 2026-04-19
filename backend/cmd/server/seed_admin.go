package main

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	pluginModel "github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	poetryModel "github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
		{AuthorityId: 10010, AuthorityName: "PluginProvider", ParentId: 0, DefaultRouter: "plugin/project-management"},
		{AuthorityId: 10013, AuthorityName: "PluginReviewer", ParentId: 0, DefaultRouter: "plugin/work-order-pool"},
		{AuthorityId: 888, AuthorityName: "CommonUser", ParentId: 0, DefaultRouter: "dashboard/workplace"},
		{AuthorityId: 9528, AuthorityName: "TestUser", ParentId: 0, DefaultRouter: "dashboard/workplace"},
		{AuthorityId: 8881, AuthorityName: "CommonUserChild", ParentId: 888, DefaultRouter: "dashboard/workplace"},
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
		{Key: "sys_api_token", ParentKey: "sys_root", Path: "/sys/api-token", Name: "menu.system.apiToken", Component: "sys/api-token", Icon: "KeyOutlined", Sort: 5, Locale: "menu.system.apiToken"},
		{Key: "sys_operation", ParentKey: "sys_root", Path: "/sys/operation", Name: "menu.system.operation", Component: "sys/operation", Icon: "HistoryOutlined", Sort: 6, Locale: "menu.system.operation"},
		{Key: "sys_notice", ParentKey: "sys_root", Path: "/sys/notice", Name: "menu.system.notice", Component: "sys/notice", Icon: "NotificationOutlined", Sort: 7, Locale: "menu.system.notice"},
		{Key: "sys_plugin_master", ParentKey: "sys_root", Path: "/sys/plugin-master", Name: "menu.system.pluginMaster", Component: "sys/plugin-master", Icon: "DatabaseOutlined", Sort: 8, Locale: "menu.system.pluginMaster"},
		{Key: "plugin_root", Path: "/plugin", Name: "menu.plugin", Component: "components/RouterLayout", Icon: "AppstoreAddOutlined", Sort: 15, Locale: "menu.plugin"},
		{Key: "plugin_project_management", ParentKey: "plugin_root", Path: "/plugin/project-management", Name: "menu.plugin.projectManagement", Component: "plugin/project-management", Icon: "FolderOpenOutlined", Sort: 1, Locale: "menu.plugin.projectManagement"},
		{Key: "plugin_project_detail", ParentKey: "plugin_root", Path: "/plugin/project/:id", Name: "menu.plugin.projectDetail", Component: "plugin/project-detail", Icon: "ProfileOutlined", Sort: 2, Locale: "menu.plugin.projectDetail", HideInMenu: true},
		{Key: "plugin_work_order_pool", ParentKey: "plugin_root", Path: "/plugin/work-order-pool", Name: "menu.plugin.workOrderPool", Component: "plugin/work-order-pool", Icon: "AuditOutlined", Sort: 3, Locale: "menu.plugin.workOrderPool"},
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
		adminAuthorityID: {"dashboard", "state", "about", "sys_root", "sys_user", "sys_authority", "sys_menu", "sys_api", "sys_api_token", "sys_operation", "sys_notice", "sys_plugin_master", "plugin_root", "plugin_project_management", "plugin_project_detail", "plugin_work_order_pool", "poetry_root", "poetry_dynasty", "poetry_genre", "poetry_author", "poetry_poem", "account_settings"},
		10010:            {"dashboard", "about", "plugin_root", "plugin_project_management", "plugin_project_detail", "account_settings"},
		10013:            {"dashboard", "about", "plugin_root", "plugin_project_detail", "plugin_work_order_pool", "account_settings"},
		9528:             {"dashboard", "state", "about", "sys_root", "sys_user", "sys_authority", "sys_menu", "sys_api", "sys_api_token", "sys_operation", "sys_notice", "sys_plugin_master", "plugin_root", "plugin_project_management", "plugin_project_detail", "plugin_work_order_pool", "poetry_root", "poetry_dynasty", "poetry_genre", "poetry_author", "poetry_poem", "account_settings"},
		888:              {"dashboard", "state", "about", "poetry_root", "poetry_dynasty", "poetry_genre", "poetry_author", "poetry_poem", "account_settings"},
		8881:             {"dashboard", "about", "account_settings"},
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

		{Path: "/api/v1/sys/api-token/getApiTokenList", Method: "POST", ApiGroup: "system-api-token", Description: "Get API token list"},
		{Path: "/api/v1/sys/api-token/detail", Method: "GET", ApiGroup: "system-api-token", Description: "Get API token detail"},
		{Path: "/api/v1/sys/api-token/create", Method: "POST", ApiGroup: "system-api-token", Description: "Create API token"},
		{Path: "/api/v1/sys/api-token/update", Method: "PUT", ApiGroup: "system-api-token", Description: "Update API token"},
		{Path: "/api/v1/sys/api-token/delete", Method: "DELETE", ApiGroup: "system-api-token", Description: "Delete API token"},
		{Path: "/api/v1/sys/api-token/reset", Method: "POST", ApiGroup: "system-api-token", Description: "Reset API token"},
		{Path: "/api/v1/sys/api-token/enable", Method: "POST", ApiGroup: "system-api-token", Description: "Enable API token"},
		{Path: "/api/v1/sys/api-token/disable", Method: "POST", ApiGroup: "system-api-token", Description: "Disable API token"},

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
		{Path: "/api/v1/plugin/plugin/getProjectDetail", Method: "POST", ApiGroup: "plugin", Description: "Get plugin project detail"},
		{Path: "/api/v1/plugin/plugin/createPlugin", Method: "POST", ApiGroup: "plugin", Description: "Create plugin"},
		{Path: "/api/v1/plugin/plugin/updatePlugin", Method: "PUT", ApiGroup: "plugin", Description: "Update plugin"},
		{Path: "/api/v1/plugin/release/getReleaseDetail", Method: "POST", ApiGroup: "plugin", Description: "Get release detail"},
		{Path: "/api/v1/plugin/release/createRelease", Method: "POST", ApiGroup: "plugin", Description: "Create release"},
		{Path: "/api/v1/plugin/release/updateRelease", Method: "PUT", ApiGroup: "plugin", Description: "Update release"},
		{Path: "/api/v1/plugin/release/transition", Method: "POST", ApiGroup: "plugin", Description: "Transition release"},
		{Path: "/api/v1/plugin/release/claim", Method: "POST", ApiGroup: "plugin", Description: "Claim release work order"},
		{Path: "/api/v1/plugin/release/reset", Method: "POST", ApiGroup: "plugin", Description: "Reset release work order"},
		{Path: "/api/v1/plugin/work-order/getWorkOrderPool", Method: "POST", ApiGroup: "plugin", Description: "Get work order pool"},
		{Path: "/api/v1/plugin/product/getProductList", Method: "POST", ApiGroup: "plugin", Description: "Get product list"},
		{Path: "/api/v1/plugin/product/createProduct", Method: "POST", ApiGroup: "plugin", Description: "Create product"},
		{Path: "/api/v1/plugin/product/updateProduct", Method: "PUT", ApiGroup: "plugin", Description: "Update product"},
		{Path: "/api/v1/plugin/department/getDepartmentList", Method: "POST", ApiGroup: "plugin", Description: "Get department list"},
		{Path: "/api/v1/plugin/department/createDepartment", Method: "POST", ApiGroup: "plugin", Description: "Create department"},
		{Path: "/api/v1/plugin/department/updateDepartment", Method: "PUT", ApiGroup: "plugin", Description: "Update department"},
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
		apiSign("POST", "/api/v1/sys/api-token/getApiTokenList"),
		apiSign("GET", "/api/v1/sys/api-token/detail"),
		apiSign("POST", "/api/v1/sys/api-token/create"),
		apiSign("PUT", "/api/v1/sys/api-token/update"),
		apiSign("DELETE", "/api/v1/sys/api-token/delete"),
		apiSign("POST", "/api/v1/sys/api-token/reset"),
		apiSign("POST", "/api/v1/sys/api-token/enable"),
		apiSign("POST", "/api/v1/sys/api-token/disable"),
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
		apiSign("POST", "/api/v1/plugin/release/getReleaseDetail"),
		apiSign("POST", "/api/v1/plugin/release/createRelease"),
		apiSign("PUT", "/api/v1/plugin/release/updateRelease"),
		apiSign("POST", "/api/v1/plugin/release/transition"),
		apiSign("POST", "/api/v1/plugin/release/claim"),
		apiSign("POST", "/api/v1/plugin/release/reset"),
		apiSign("POST", "/api/v1/plugin/work-order/getWorkOrderPool"),
		apiSign("POST", "/api/v1/plugin/product/getProductList"),
		apiSign("POST", "/api/v1/plugin/product/createProduct"),
		apiSign("PUT", "/api/v1/plugin/product/updateProduct"),
		apiSign("POST", "/api/v1/plugin/department/getDepartmentList"),
		apiSign("POST", "/api/v1/plugin/public/getPublishedPluginList"),
		apiSign("POST", "/api/v1/plugin/public/getPublishedPluginDetail"),
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
		10010: append(basicAccess,
			apiSign("POST", "/api/v1/plugin/plugin/getPluginList"),
			apiSign("POST", "/api/v1/plugin/plugin/getProjectDetail"),
			apiSign("POST", "/api/v1/plugin/plugin/createPlugin"),
			apiSign("PUT", "/api/v1/plugin/plugin/updatePlugin"),
			apiSign("POST", "/api/v1/plugin/release/getReleaseDetail"),
			apiSign("POST", "/api/v1/plugin/release/createRelease"),
			apiSign("PUT", "/api/v1/plugin/release/updateRelease"),
			apiSign("POST", "/api/v1/plugin/release/transition"),
			apiSign("POST", "/api/v1/plugin/product/getProductList"),
			apiSign("POST", "/api/v1/plugin/department/getDepartmentList"),
			apiSign("POST", "/api/v1/plugin/public/getPublishedPluginList"),
			apiSign("POST", "/api/v1/plugin/public/getPublishedPluginDetail"),
		),
		10013: append(basicAccess,
			apiSign("POST", "/api/v1/plugin/plugin/getProjectDetail"),
			apiSign("POST", "/api/v1/plugin/release/getReleaseDetail"),
			apiSign("POST", "/api/v1/plugin/release/transition"),
			apiSign("POST", "/api/v1/plugin/release/claim"),
			apiSign("POST", "/api/v1/plugin/work-order/getWorkOrderPool"),
			apiSign("POST", "/api/v1/plugin/product/getProductList"),
			apiSign("POST", "/api/v1/plugin/department/getDepartmentList"),
			apiSign("POST", "/api/v1/plugin/public/getPublishedPluginList"),
			apiSign("POST", "/api/v1/plugin/public/getPublishedPluginDetail"),
		),
		9528: fullAccess,
		888:  basicAccess,
		8881: basicAccess,
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
		{"POST", "/api/v1/sys/api-token/getApiTokenList"},
		{"GET", "/api/v1/sys/api-token/detail"},
		{"POST", "/api/v1/sys/api-token/create"},
		{"PUT", "/api/v1/sys/api-token/update"},
		{"DELETE", "/api/v1/sys/api-token/delete"},
		{"POST", "/api/v1/sys/api-token/reset"},
		{"POST", "/api/v1/sys/api-token/enable"},
		{"POST", "/api/v1/sys/api-token/disable"},
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
		{"POST", "/api/v1/plugin/plugin/getProjectDetail"},
		{"POST", "/api/v1/plugin/plugin/createPlugin"},
		{"PUT", "/api/v1/plugin/plugin/updatePlugin"},
		{"POST", "/api/v1/plugin/release/getReleaseDetail"},
		{"POST", "/api/v1/plugin/release/createRelease"},
		{"PUT", "/api/v1/plugin/release/updateRelease"},
		{"POST", "/api/v1/plugin/release/transition"},
		{"POST", "/api/v1/plugin/release/claim"},
		{"POST", "/api/v1/plugin/release/reset"},
		{"POST", "/api/v1/plugin/work-order/getWorkOrderPool"},
		{"POST", "/api/v1/plugin/product/getProductList"},
		{"POST", "/api/v1/plugin/product/createProduct"},
		{"PUT", "/api/v1/plugin/product/updateProduct"},
		{"POST", "/api/v1/plugin/department/getDepartmentList"},
		{"POST", "/api/v1/plugin/public/getPublishedPluginList"},
		{"POST", "/api/v1/plugin/public/getPublishedPluginDetail"},
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
		10010: append(basicAccess,
			[]string{"POST", "/api/v1/plugin/plugin/getPluginList"},
			[]string{"POST", "/api/v1/plugin/plugin/getProjectDetail"},
			[]string{"POST", "/api/v1/plugin/plugin/createPlugin"},
			[]string{"PUT", "/api/v1/plugin/plugin/updatePlugin"},
			[]string{"POST", "/api/v1/plugin/release/getReleaseDetail"},
			[]string{"POST", "/api/v1/plugin/release/createRelease"},
			[]string{"PUT", "/api/v1/plugin/release/updateRelease"},
			[]string{"POST", "/api/v1/plugin/release/transition"},
			[]string{"POST", "/api/v1/plugin/product/getProductList"},
			[]string{"POST", "/api/v1/plugin/department/getDepartmentList"},
			[]string{"POST", "/api/v1/plugin/public/getPublishedPluginList"},
			[]string{"POST", "/api/v1/plugin/public/getPublishedPluginDetail"},
		),
		10013: append(basicAccess,
			[]string{"POST", "/api/v1/plugin/plugin/getProjectDetail"},
			[]string{"POST", "/api/v1/plugin/release/getReleaseDetail"},
			[]string{"POST", "/api/v1/plugin/release/transition"},
			[]string{"POST", "/api/v1/plugin/release/claim"},
			[]string{"POST", "/api/v1/plugin/work-order/getWorkOrderPool"},
			[]string{"POST", "/api/v1/plugin/product/getProductList"},
			[]string{"POST", "/api/v1/plugin/department/getDepartmentList"},
			[]string{"POST", "/api/v1/plugin/public/getPublishedPluginList"},
			[]string{"POST", "/api/v1/plugin/public/getPublishedPluginDetail"},
		),
		9528: fullAccess,
		888:  basicAccess,
		8881: basicAccess,
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

func ensurePoetryBaseData(tx *gorm.DB) error {
	dynasties, err := ensureDynasties(tx)
	if err != nil {
		return err
	}
	if err := ensurePluginBaseDataClean(tx); err != nil {
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

func ensurePluginBaseData(tx *gorm.DB) error {
	departments := []pluginModel.PluginDepartment{
		{Name: "存储产品部", NameZh: "存储产品部", NameEn: "Storage Products", ProductLine: "Storage", Sort: 1, Status: true},
		{Name: "网络产品部", NameZh: "网络产品部", NameEn: "Network Products", ProductLine: "Network", Sort: 2, Status: true},
	}
	for _, seed := range departments {
		var item pluginModel.PluginDepartment
		err := tx.Where("name = ?", seed.Name).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&seed).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&item).Updates(map[string]interface{}{
			"name_zh":      seed.NameZh,
			"name_en":      seed.NameEn,
			"product_line": seed.ProductLine,
			"sort":         seed.Sort,
			"status":       seed.Status,
		}).Error; err != nil {
			return err
		}
	}

	products := []pluginModel.PluginProduct{
		{Code: "HCI", Name: "超融合", Type: pluginModel.CompatibleTargetTypeProduct, Description: "HCI 平台", Sort: 1, Status: true},
		{Code: "ACLI", Name: "aCLI", Type: pluginModel.CompatibleTargetTypeAcli, Description: "aCLI 命令行工具", Sort: 2, Status: true},
		{Code: "VDC", Name: "云桌面", Type: pluginModel.CompatibleTargetTypeProduct, Description: "VDC 平台", Sort: 3, Status: true},
	}
	for _, seed := range products {
		var item pluginModel.PluginProduct
		err := tx.Where("code = ?", seed.Code).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&seed).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&item).Updates(map[string]interface{}{
			"name":        seed.Name,
			"type":        seed.Type,
			"description": seed.Description,
			"sort":        seed.Sort,
			"status":      seed.Status,
		}).Error; err != nil {
			return err
		}
	}
	return nil
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

func ensurePluginBaseDataClean(tx *gorm.DB) error {
	departments := []pluginModel.PluginDepartment{
		{Name: "存储产品部", NameZh: "存储产品部", NameEn: "Storage Products", ProductLine: "Storage", Sort: 1, Status: true},
		{Name: "网络产品部", NameZh: "网络产品部", NameEn: "Network Products", ProductLine: "Network", Sort: 2, Status: true},
	}
	for _, seed := range departments {
		var item pluginModel.PluginDepartment
		err := tx.Where("sort = ?", seed.Sort).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&seed).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&item).Updates(map[string]interface{}{
			"name":         seed.Name,
			"name_zh":      seed.NameZh,
			"name_en":      seed.NameEn,
			"product_line": seed.ProductLine,
			"sort":         seed.Sort,
			"status":       seed.Status,
		}).Error; err != nil {
			return err
		}
	}

	products := []pluginModel.PluginProduct{
		{Code: "HCI", Name: "超融合", Type: pluginModel.CompatibleTargetTypeProduct, Description: "HCI 平台", Sort: 1, Status: true},
		{Code: "ACLI", Name: "aCLI", Type: pluginModel.CompatibleTargetTypeAcli, Description: "aCLI 命令行工具", Sort: 2, Status: true},
		{Code: "VDC", Name: "云桌面", Type: pluginModel.CompatibleTargetTypeProduct, Description: "VDC 平台", Sort: 3, Status: true},
	}
	for _, seed := range products {
		var item pluginModel.PluginProduct
		err := tx.Where("code = ?", seed.Code).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&seed).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&item).Updates(map[string]interface{}{
			"name":        seed.Name,
			"type":        seed.Type,
			"description": seed.Description,
			"sort":        seed.Sort,
			"status":      seed.Status,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
