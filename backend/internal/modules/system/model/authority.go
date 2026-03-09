package model

import "time"

type SysAuthority struct {
	CreatedAt time.Time  `json:"createdAt"` // 创建时间
	UpdatedAt time.Time  `json:"updatedAt"` // 更新时间
	DeletedAt *time.Time `json:"-" sql:"index"`

	// GVA 特色：允许自定义 ID (非自增)，所以这里不使用 common.BaseModel
	AuthorityId   uint   `json:"authorityId" gorm:"not null;unique;primary_key;comment:角色ID;size:90"`
	AuthorityName string `json:"authorityName" gorm:"comment:角色名"`
	ParentId      uint   `json:"parentId" gorm:"default:0;comment:父角色ID"` // 推荐使用 uint default 0 而非指针
	DefaultRouter string `json:"defaultRouter" gorm:"comment:默认菜单;default:dashboard"`

	// 数据权限 (多对多自关联)
	DataAuthorityId []*SysAuthority `json:"dataAuthorityId" gorm:"many2many:sys_data_authority_id;"`

	// 子角色 (树形结构)
	Children []SysAuthority `json:"children" gorm:"-"`

	// 多对多关联：角色 <-> 菜单
	// 关联表: sys_authority_menus
	// joinForeignKey: 当前表在关联表中的列 (authority_id)
	// joinReferences: 对方表在关联表中的列 (menu_id)
	SysMenus []SysMenu `json:"menus" gorm:"many2many:sys_authority_menus;joinForeignKey:authority_id;joinReferences:menu_id;"`
	// 2. 角色 <-> API (多对多)
	// 关联表: sys_authority_apis
	// joinForeignKey: authority_id (当前表在关联表中的列)
	// joinReferences: api_id (对方表在关联表中的列)
	SysApis []SysApi `json:"apis" gorm:"many2many:sys_authority_apis;joinForeignKey:authority_id;joinReferences:api_id;"`

	// 多对多关联：角色 <-> 用户 (反向查询用)
	Users []SysUser `json:"-" gorm:"many2many:sys_user_authorities;joinForeignKey:authority_id;joinReferences:user_id;"`
}

func (SysAuthority) TableName() string {
	return "sys_authorities"
}
