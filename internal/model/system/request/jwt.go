package request

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomClaims 自定义荷载
type CustomClaims struct {
	BaseClaims
	BufferTime int64 `json:"bufferTime"` // 缓冲时间 (秒)
	jwt.RegisteredClaims
}

type BaseClaims struct {
	UUID        uuid.UUID `json:"uuid"`
	UserID      uint      `json:"userId"`
	Username    string    `json:"username"`
	NickName    string    `json:"nickName"`
	AuthorityId uint      `json:"authorityId"`
}
