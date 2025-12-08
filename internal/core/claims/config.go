package claims

import (
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义荷载
type CustomClaims struct {
	dto.BaseClaims
	BufferTime int64 `json:"bufferTime"` // 缓冲时间 (秒)
	jwt.RegisteredClaims
}
