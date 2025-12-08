package dto

// CasbinInfo Casbin 规则详情
type CasbinInfo struct {
	Path   string `json:"path"`   // API 路径
	Method string `json:"method"` // 请求方法
}

// UpdateCasbinReq 更新角色API权限请求
type UpdateCasbinReq struct {
	AuthorityId string       `json:"authorityId" binding:"required"` // 角色ID
	CasbinInfos []CasbinInfo `json:"casbinInfos"`                    // 选中的 API 列表
}

// GetPolicyPathByAuthorityIdReq 获取角色当前拥有的API权限
type GetPolicyPathByAuthorityIdReq struct {
	AuthorityId string `json:"authorityId" binding:"required"`
}
