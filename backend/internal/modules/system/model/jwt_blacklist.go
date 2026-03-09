package model

import "github.com/CIPFZ/gowebframe/internal/modules/common"

type JwtBlacklist struct {
	common.BaseModel
	Jwt string `gorm:"type:text;comment:jwt"`
}
