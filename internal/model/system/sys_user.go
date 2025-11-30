package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

//type Login interface {
//	GetUsername() string
//	GetNickname() string
//	GetUUID() uuid.UUID
//	GetUserId() uint
//	GetAuthorityId() uint
//	GetUserInfo() any
//}
//
//var _ Login = new(SysUser)
//
//type SysUser struct {
//	common.BaseModel
//	UUID          uuid.UUID      `json:"uuid" gorm:"index;comment:用户UUID"`                                                                   // 用户UUID
//	Username      string         `json:"userName" gorm:"index;comment:用户登录名"`                                                                // 用户登录名
//	Password      string         `json:"-"  gorm:"comment:用户登录密码"`                                                                           // 用户登录密码
//	NickName      string         `json:"nickName" gorm:"default:系统用户;comment:用户昵称"`                                                          // 用户昵称
//	HeaderImg     string         `json:"headerImg" gorm:"default:https://qmplusimg.henrongyi.top/gva_header.jpg;comment:用户头像"`               // 用户头像
//	AuthorityId   uint           `json:"authorityId" gorm:"default:888;comment:用户角色ID"`                                                      // 用户角色ID
//	Authority     SysAuthority   `json:"authority" gorm:"foreignKey:AuthorityId;references:AuthorityId;comment:用户角色"`                        // 用户角色
//	Authorities   []SysAuthority `json:"authorities" gorm:"many2many:sys_user_authority;"`                                                   // 多用户角色
//	Phone         string         `json:"phone"  gorm:"comment:用户手机号"`                                                                        // 用户手机号
//	Email         string         `json:"email"  gorm:"comment:用户邮箱"`                                                                         // 用户邮箱
//	Enable        int            `json:"enable" gorm:"default:1;comment:用户是否被冻结 1正常 2冻结"`                                                    //用户是否被冻结 1正常 2冻结
//	OriginSetting common.JSONMap `json:"originSetting" form:"originSetting" gorm:"type:text;default:null;column:origin_setting;comment:配置;"` //配置
//}
//
//func (SysUser) TableName() string {
//	return "sys_users"
//}
//
//func (s *SysUser) GetUsername() string {
//	return s.Username
//}
//
//func (s *SysUser) GetNickname() string {
//	return s.NickName
//}
//
//func (s *SysUser) GetUUID() uuid.UUID {
//	return s.UUID
//}
//
//func (s *SysUser) GetUserId() uint {
//	return s.ID
//}
//
//func (s *SysUser) GetAuthorityId() uint {
//	return s.AuthorityId
//}
//
//func (s *SysUser) GetUserInfo() any {
//	return *s
//}

type SysUser struct {
	common.BaseModel // 包含 ID, CreatedAt, UpdatedAt, DeletedAt

	// --- 身份认证 ---
	UUID     uuid.UUID `json:"uuid" gorm:"type:char(36);index;comment:用户UUID"`
	Username string    `json:"username" gorm:"type:varchar(64);uniqueIndex;comment:用户名"`
	Password string    `json:"-" gorm:"type:varchar(128);comment:密码"` // JSON 隐藏

	// --- 个人信息 ---
	NickName string `json:"nickName" gorm:"type:varchar(64);default:系统用户;comment:昵称"`
	Avatar   string `json:"avatar" gorm:"type:varchar(255);default:https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png;comment:头像"`
	Phone    string `json:"phone" gorm:"type:varchar(20);comment:手机号"`
	Email    string `json:"email" gorm:"type:varchar(128);comment:邮箱"`

	// --- 状态与配置 ---
	Status   int            `json:"status" gorm:"type:tinyint(1);default:1;comment:用户状态 1正常 2冻结"`
	Settings datatypes.JSON `json:"settings" gorm:"type:json;comment:个性化设置"`

	// --- 权限关联 ---
	AuthorityID uint         `json:"authorityId" gorm:"default:888;comment:当前角色ID"`
	Authority   SysAuthority `json:"authority" gorm:"foreignKey:AuthorityID;references:AuthorityId;comment:当前角色"`

	// 多对多关联：用户 <-> 角色
	// 关联表: sys_user_authorities
	// joinForeignKey: 当前表在关联表中的列名 (user_id)
	// joinReferences: 对方表在关联表中的列名 (authority_id)
	Authorities []SysAuthority `json:"authorities" gorm:"many2many:sys_user_authorities;joinForeignKey:user_id;joinReferences:authority_id;"`
}

func (SysUser) TableName() string {
	return "sys_users"
}
