package model

// SysAuthorityApi 角色-API 关联表 (Join Table)
type SysAuthorityApi struct {
	AuthorityId uint `gorm:"column:authority_id;primaryKey;comment:角色ID"`
	ApiId       uint `gorm:"column:api_id;primaryKey;comment:API ID"`
}

func (SysAuthorityApi) TableName() string {
	return "sys_authority_apis"
}
