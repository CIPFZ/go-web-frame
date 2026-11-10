package model

import (
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
)

type SysApiToken struct {
	common.BaseModel

	TokenHash      string     `json:"-" gorm:"type:varchar(64);uniqueIndex;comment:token hash"`
	TokenPrefix    string     `json:"tokenPrefix" gorm:"type:varchar(16);comment:token prefix"`
	Name           string     `json:"name" gorm:"type:varchar(100);comment:token name"`
	Description    string     `json:"description" gorm:"type:varchar(255);comment:token description"`
	ExpiresAt      *time.Time `json:"expiresAt" gorm:"comment:expires at"`
	MaxConcurrency int        `json:"maxConcurrency" gorm:"default:5;comment:max concurrency"`
	Enabled        bool       `json:"enabled" gorm:"default:true;comment:enabled"`
	LastUsedAt     *time.Time `json:"lastUsedAt" gorm:"comment:last used at"`
	CreatedBy      uint       `json:"createdBy" gorm:"comment:token owner"` // 主人：Token 跟主人绑定，只有本人（或管理员）能管理
	// FollowOwner：权限跟随主人的角色（可被 Token 调用的接口中主人有权限的）；否则只能调用 Apis 里选的接口。两种情况都不会超出主人现在的权限
	FollowOwner bool `json:"followOwner" gorm:"not null;default:false;comment:follow owner permissions"`

	Apis []SysApi `json:"apis" gorm:"many2many:sys_api_token_apis;joinForeignKey:ApiTokenId;joinReferences:ApiId"`
}

func (SysApiToken) TableName() string {
	return "sys_api_tokens"
}

type SysApiTokenApi struct {
	ApiTokenId uint `gorm:"column:api_token_id;primaryKey;comment:api token id"`
	ApiId      uint `gorm:"column:api_id;primaryKey;comment:api id"`
}

func (SysApiTokenApi) TableName() string {
	return "sys_api_token_apis"
}

// SysTokenLimit 是 Token 申请限制：UserID 为 0 的一行是全局默认，其余是单个用户的设置；字段为空表示跟随全局默认。
type SysTokenLimit struct {
	UserID         uint `json:"userId" gorm:"primaryKey;autoIncrement:false;comment:user id, 0 is the global default"`
	MaxTokens      *int `json:"maxTokens" gorm:"comment:max tokens per user"`
	MaxTTLDays     *int `json:"maxTtlDays" gorm:"column:max_ttl_days;comment:max validity days"`
	MaxConcurrency *int `json:"maxConcurrency" gorm:"comment:max concurrency per token"`
}

func (SysTokenLimit) TableName() string { return "sys_token_limits" }
