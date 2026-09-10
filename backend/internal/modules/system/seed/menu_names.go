package seed

// MenuNames separates the editable display name from its translation key.
var MenuNames = map[string]string{
	"menu.dashboard.workplace": "工作台",
	"menu.state":               "系统状态",
	"menu.about":               "关于",
	"menu.system":              "系统管理",
	"menu.system.user":         "用户管理",
	"menu.system.authority":    "角色管理",
	"menu.system.menu":         "菜单管理",
	"menu.system.api":          "API 管理",
	"menu.system.apiToken":     "API Token",
	"menu.system.operation":    "操作日志",
	"menu.system.notice":       "通知公告",
	"menu.account.settings":    "个人设置",
}

// MenuNamesEn is shared by new installations and the additive locale migration.
var MenuNamesEn = map[string]string{
	"menu.dashboard.workplace": "Workplace", "menu.state": "System status", "menu.about": "About",
	"menu.system": "System", "menu.system.user": "Users", "menu.system.authority": "Roles",
	"menu.system.menu": "Menus", "menu.system.api": "APIs", "menu.system.apiToken": "API Tokens",
	"menu.system.operation": "Operation Logs", "menu.system.notice": "Notices", "menu.account.settings": "Account settings",
}
