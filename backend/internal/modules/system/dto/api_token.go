package dto

import "github.com/CIPFZ/gowebframe/internal/modules/common"

type SearchApiTokenReq struct {
	Name    string `json:"name" form:"name"`
	Enabled *bool  `json:"enabled" form:"enabled"`
	OwnerID uint   `json:"ownerId" form:"ownerId"` // 只看某个人的 Token
	common.PageInfo
}

type CreateApiTokenReq struct {
	Name           string  `json:"name" binding:"required,max=100"`
	Description    string  `json:"description" binding:"max=255"`
	ExpiresAt      *string `json:"expiresAt"` // 为空时按本人的最长有效期
	NeverExpire    bool    `json:"neverExpire"`
	MaxConcurrency int     `json:"maxConcurrency" binding:"omitempty,min=1,max=1000"` // 为空时按本人的上限
	FollowOwner    bool    `json:"followOwner"`                                       // 权限跟随本人的角色；否则用 ApiIds
	ApiIds         []uint  `json:"apiIds"`
}

type UpdateApiTokenReq struct {
	ID             uint    `json:"id" binding:"required"`
	Name           string  `json:"name" binding:"required,max=100"`
	Description    string  `json:"description" binding:"max=255"`
	ExpiresAt      *string `json:"expiresAt"`
	NeverExpire    bool    `json:"neverExpire"`
	MaxConcurrency int     `json:"maxConcurrency" binding:"omitempty,min=1,max=1000"` // 为空时保持原值
	FollowOwner    bool    `json:"followOwner"`                                       // 权限跟随本人的角色；否则用 ApiIds
	ApiIds         []uint  `json:"apiIds"`
}

type DeleteApiTokenReq struct {
	ID  uint   `json:"id"`
	IDs []uint `json:"ids"`
}

type ToggleApiTokenReq struct {
	ID uint `json:"id" binding:"required"`
}

type ApiTokenDetailReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

type ApiTokenResponse struct {
	ID             uint            `json:"ID"`
	TokenPrefix    string          `json:"tokenPrefix"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ExpiresAt      *string         `json:"expiresAt"`
	NeverExpire    bool            `json:"neverExpire"`
	MaxConcurrency int             `json:"maxConcurrency"`
	Enabled        bool            `json:"enabled"`
	LastUsedAt     *string         `json:"lastUsedAt"`
	CreatedAt      string          `json:"createdAt"`
	CreatedBy      uint            `json:"createdBy"`
	OwnerName      string          `json:"ownerName"`
	FollowOwner    bool            `json:"followOwner"`
	Apis           []ApiSimpleItem `json:"apis"`
}

type ApiTokenSecretResponse struct {
	ApiTokenResponse
	Token string `json:"token"`
}

type ApiSimpleItem struct {
	ID          uint   `json:"ID"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	ApiGroup    string `json:"apiGroup"`
	Description string `json:"description"`
}

// TokenLimits 是生效的 Token 申请限制。
type TokenLimits struct {
	MaxTokens      int `json:"maxTokens"`
	MaxTTLDays     int `json:"maxTtlDays"`
	MaxConcurrency int `json:"maxConcurrency"`
}

// TokenLimitSettings 是一份限制设置（UserID 为 0 表示全局默认）；字段为空表示跟随默认。
type TokenLimitSettings struct {
	UserID         uint `json:"userId"`
	MaxTokens      *int `json:"maxTokens" binding:"omitempty,min=0,max=1000"`
	MaxTTLDays     *int `json:"maxTtlDays" binding:"omitempty,min=1,max=3650"`
	MaxConcurrency *int `json:"maxConcurrency" binding:"omitempty,min=1,max=1000"`
}

// TokenLimitDetail 是某个用户的设置、实际生效的值与全局默认。
type TokenLimitDetail struct {
	TokenLimitSettings
	Effective TokenLimits `json:"effective"`
	Default   TokenLimits `json:"default"`
}

type TokenLimitReq struct {
	UserID uint `json:"userId"`
}
