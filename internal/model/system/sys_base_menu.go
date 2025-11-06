package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common"
)

//type SysBaseMenu struct {
//	common.BaseModel
//	MenuLevel     uint                                       `json:"-"`
//	ParentId      uint                                       `json:"parentId" gorm:"comment:父菜单ID"`     // 父菜单ID
//	Path          string                                     `json:"path" gorm:"comment:路由path"`        // 路由path
//	Name          string                                     `json:"name" gorm:"comment:路由name"`        // 路由name
//	Component     string                                     `json:"component" gorm:"comment:对应前端文件路径"` // 对应前端文件路径
//	Locale        string                                     `json:"locale" gorm:"comment:国际化"`
//	Hidden        bool                                       `json:"hidden" gorm:"comment:是否在列表隐藏"` // 是否在列表隐藏
//	Sort          int                                        `json:"sort" gorm:"comment:排序标记"`      // 排序标记
//	Meta          `json:"meta" gorm:"embedded;comment:附加属性"` // 附加属性
//	SysAuthoritys []SysAuthority                             `json:"authoritys" gorm:"many2many:sys_authority_menus;"`
//	Children      []SysBaseMenu                              `json:"children" gorm:"-"`
//	Parameters    []SysBaseMenuParameter                     `json:"parameters"`
//	MenuBtn       []SysBaseMenuBtn                           `json:"menuBtn"`
//}

// SysBaseMenu 适配 Ant Design Pro 的新 GORM 结构
type SysBaseMenu struct {
	common.BaseModel // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
	// --- 核心结构 ---
	ParentId uint `json:"parentId" gorm:"default:0;comment:父菜单ID"` // 父菜单ID (0为根)
	Sort     int  `json:"sort" gorm:"default:0;comment:排序标记"`      // 排序标记
	// --- Ant Design Pro 核心字段 ---
	// (GVA 的 'title' 已被 'name' 取代)
	Name      string `json:"name" gorm:"comment:菜单显示名称 (e.g., 仪表盘)"`
	Path      string `json:"path" gorm:"comment:路由path (e.g., /dashboard)"`
	Icon      string `json:"icon" gorm:"comment:Antd图标名 (e.g., DashboardOutlined)"`
	Component string `json:"component" gorm:"comment:React组件路径 (e.g., Dashboard, Admin/User)"`
	// --- Ant Design Pro 扩展功能 ---
	// (GVA 的 'hidden' 已被 'hideInMenu' 取代)
	HideInMenu bool   `json:"hideInMenu" gorm:"default:0;comment:是否在菜单中隐藏"`
	Access     string `json:"access" gorm:"comment:权限标识 (e.g., canAdmin)"` // 对应前端 access.ts
	Target     string `json:"target" gorm:"comment:外链target (_blank)"`     // e.g., _blank
	Locale     string `json:"locale" gorm:"comment:国际化key (e.g., menu.dashboard)"`
	// --- 关系 ---
	// SysAuthoritys 用于 GORM 查询，但在 JSON 响应中隐藏
	SysAuthoritys []SysAuthority `json:"-" gorm:"many2many:sys_authority_menus;"`
	// 子菜单 (用于 List-to-Tree 后返回给前端)
	// 关键: JSON 标签为 "routes" 以同时适配菜单和路由
	Children []SysBaseMenu `json:"routes" gorm:"-"`
}

type Meta struct {
	ActiveName     string `json:"activeName" gorm:"comment:高亮菜单"`
	KeepAlive      bool   `json:"keepAlive" gorm:"comment:是否缓存"`           // 是否缓存
	DefaultMenu    bool   `json:"defaultMenu" gorm:"comment:是否是基础路由（开发中）"` // 是否是基础路由（开发中）
	Title          string `json:"title" gorm:"comment:菜单名"`                // 菜单名
	Icon           string `json:"icon" gorm:"comment:菜单图标"`                // 菜单图标
	CloseTab       bool   `json:"closeTab" gorm:"comment:自动关闭tab"`         // 自动关闭tab
	TransitionType string `json:"transitionType" gorm:"comment:路由切换动画"`    // 路由切换动画
}

type SysBaseMenuParameter struct {
	common.BaseModel
	SysBaseMenuID uint
	Type          string `json:"type" gorm:"comment:地址栏携带参数为params还是query"` // 地址栏携带参数为params还是query
	Key           string `json:"key" gorm:"comment:地址栏携带参数的key"`            // 地址栏携带参数的key
	Value         string `json:"value" gorm:"comment:地址栏携带参数的值"`            // 地址栏携带参数的值
}

func (SysBaseMenu) TableName() string {
	return "sys_base_menus"
}
