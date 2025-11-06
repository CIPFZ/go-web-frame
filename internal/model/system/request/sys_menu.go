package request

import (
	"github.com/CIPFZ/gowebframe/internal/model/common"
	"github.com/CIPFZ/gowebframe/internal/model/system"
)

// AddMenuAuthorityInfo Add menu authority info structure
type AddMenuAuthorityInfo struct {
	Menus       []system.SysBaseMenu `json:"menus"`
	AuthorityId uint                 `json:"authorityId"` // 角色ID
}

func DefaultMenu() []system.SysBaseMenu {
	return []system.SysBaseMenu{{
		BaseModel: common.BaseModel{ID: 1},
		ParentId:  0,
		Path:      "dashboard",
		Name:      "dashboard",
		Component: "view/dashboard/index.vue",
		Sort:      1,
		//Meta: system.Meta{
		//	Title: "仪表盘",
		//	Icon:  "setting",
		//},
	}}
}
