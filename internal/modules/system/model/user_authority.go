package model

// SysUserAuthority 用户-角色 关联表
type SysUserAuthority struct {
	UserId      uint `gorm:"column:user_id;primaryKey"`      // 必须是 uint
	AuthorityId uint `gorm:"column:authority_id;primaryKey"` // 必须是 uint
}

func (SysUserAuthority) TableName() string {
	return "sys_user_authorities"
}
