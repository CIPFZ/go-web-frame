package dto

// AddMenuReq 新增菜单参数
type AddMenuReq struct {
	ParentId   uint   `json:"parentId"` // 父菜单ID
	Path       string `json:"path" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Component  string `json:"component" binding:"required"`
	Sort       int    `json:"sort"`
	Icon       string `json:"icon"`
	HideInMenu bool   `json:"hideInMenu"`
	Access     string `json:"access"` // 权限标识
	Target     string `json:"target"`
	Locale     string `json:"locale"`
}

// UpdateMenuReq 更新菜单参数 (包含 ID)
type UpdateMenuReq struct {
	ID         uint   `json:"id" binding:"required"`
	ParentId   uint   `json:"parentId"`
	Path       string `json:"path"`
	Name       string `json:"name"`
	Component  string `json:"component"`
	Sort       int    `json:"sort"`
	Icon       string `json:"icon"`
	HideInMenu bool   `json:"hideInMenu"`
	Access     string `json:"access"`
	Target     string `json:"target"`
	Locale     string `json:"locale"`
}

// GetAuthorityIdReq 获取角色ID的通用请求
type GetAuthorityIdReq struct {
	AuthorityId uint `json:"authorityId" form:"authorityId" binding:"required"`
}
