package model

import (
	"fmt"
	"slices"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	UserActive             = 1
	UserInactive           = 0
	DefaultUserAuthorityID = 888
	DefaultUserAvatar      = "/avatars/a01.svg"
)

// PresetAvatars 是内置头像（前端 public/avatars：a01–a12 几何图形、animal-*.svg 动物头像）；除了上传，也可以从中选一个。
var PresetAvatars = func() []string {
	out := make([]string, 0, 24)
	for i := 1; i <= 12; i++ {
		out = append(out, fmt.Sprintf("/avatars/a%02d.svg", i))
	}
	for _, a := range []string{"bear", "capybara", "cat", "dog", "fox", "koala", "otter", "panda", "penguin", "pig", "rabbit", "tiger"} {
		out = append(out, "/avatars/animal-"+a+".svg")
	}
	return out
}()

// IsPresetAvatar 判断头像是否是内置头像之一。
func IsPresetAvatar(avatar string) bool { return slices.Contains(PresetAvatars, avatar) }

type SysUser struct {
	common.BaseModel // 包含 ID, CreatedAt, UpdatedAt, DeletedAt

	// --- 身份认证 ---
	TokenVersion uint64    `json:"-" gorm:"not null;default:1"`
	UUID         uuid.UUID `json:"uuid" gorm:"type:char(36);index;comment:用户UUID"`
	Username     string    `json:"username" gorm:"type:varchar(64);uniqueIndex;comment:用户名"`
	Password     string    `json:"-" gorm:"type:varchar(128);comment:密码"` // JSON 隐藏

	// --- 个人信息 ---
	NickName string `json:"nickName" gorm:"type:varchar(64);default:系统用户;comment:昵称"`
	Avatar   string `json:"avatar" gorm:"type:varchar(255);default:/avatars/a01.svg;comment:头像"`
	Phone    string `json:"phone" gorm:"type:varchar(20);comment:手机号"`
	Email    string `json:"email" gorm:"type:varchar(128);comment:邮箱"`
	Bio      string `json:"bio" gorm:"type:varchar(255);comment:个人简介"`

	// --- 状态与配置 ---
	Status   int            `json:"status" gorm:"type:smallint;default:1;comment:用户状态 1正常 0禁用"`
	Settings datatypes.JSON `json:"settings" gorm:"type:json;comment:个性化设置"`
	// Workplace 是本人的工作台布局（小组件与应用图标），由前端维护内容
	Workplace datatypes.JSON `json:"workplace" gorm:"type:json;comment:工作台布局"`

	// --- 权限关联 ---
	AuthorityID uint         `json:"authorityId" gorm:"default:888;comment:主角色ID（决定默认首页；权限取全部角色的并集）"`
	Authority   SysAuthority `json:"authority" gorm:"belongsTo:SysAuthority;foreignKey:AuthorityID;references:AuthorityId;comment:当前角色"`

	// 多对多关联：用户 <-> 角色
	// 关联表: sys_user_authorities
	// joinForeignKey: 当前表在关联表中的列名 (user_id)
	// joinReferences: 对方表在关联表中的列名 (authority_id)
	Authorities []SysAuthority `json:"authorities" gorm:"many2many:sys_user_authorities;joinForeignKey:user_id;joinReferences:authority_id;"`
}

func (SysUser) TableName() string {
	return "sys_users"
}
