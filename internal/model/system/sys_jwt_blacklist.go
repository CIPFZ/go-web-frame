package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common"
)

type JwtBlacklist struct {
	common.BaseModel
	Jwt string `gorm:"type:text;comment:jwt"`
}
