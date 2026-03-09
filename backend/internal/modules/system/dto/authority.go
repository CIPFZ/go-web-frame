package dto

// CreateAuthorityReq 创建角色
type CreateAuthorityReq struct {
	AuthorityId   uint   `json:"authorityId" binding:"required"` // 手动指定，例如 888
	AuthorityName string `json:"authorityName" binding:"required"`
	ParentId      uint   `json:"parentId"`      // 父角色ID，0代表根角色
	DefaultRouter string `json:"defaultRouter"` // 登录后的默认跳转路由
}

// UpdateAuthorityReq 更新角色
type UpdateAuthorityReq struct {
	AuthorityId   uint   `json:"authorityId" binding:"required"` // 主键
	AuthorityName string `json:"authorityName"`
	DefaultRouter string `json:"defaultRouter"`
	// 通常不建议修改 ParentId，因为涉及复杂的树结构变更检查
}

// SetAuthorityMenusReq 设置角色菜单权限
type SetAuthorityMenusReq struct {
	AuthorityId uint   `json:"authorityId" binding:"required"`
	MenuIds     []uint `json:"menuIds"`
}
