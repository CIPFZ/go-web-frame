package model

import "time"

// SysUserIdentity 把外部登录提供方的身份（提供方 + 提供方内的唯一标识）关联到本地用户。
type SysUserIdentity struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Provider  string    `json:"provider" gorm:"type:varchar(32);not null;uniqueIndex:idx_user_identity_subject,priority:1;comment:外部登录提供方"`
	Subject   string    `json:"subject" gorm:"type:varchar(191);not null;uniqueIndex:idx_user_identity_subject,priority:2;comment:提供方内的用户标识"`
	UserID    uint      `json:"userId" gorm:"not null;index;comment:本地用户"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SysUserIdentity) TableName() string { return "sys_user_identities" }
