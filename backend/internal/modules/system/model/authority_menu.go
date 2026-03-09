package model

// SysAuthorityMenu 角色-菜单 关联表
type SysAuthorityMenu struct {
	MenuId      uint `json:"menuId" gorm:"column:menu_id;primaryKey;comment:菜单ID"` // ✨ 修正: 必须是 uint
	AuthorityId uint `json:"-" gorm:"column:authority_id;primaryKey;comment:角色ID"` // ✨ 修正: 必须是 uint
}

func (SysAuthorityMenu) TableName() string {
	return "sys_authority_menus"
}
