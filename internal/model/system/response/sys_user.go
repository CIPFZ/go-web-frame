package response

import "github.com/CIPFZ/gowebframe/internal/model/system"

type SysUserResponse struct {
	User system.SysUser `json:"user"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}
