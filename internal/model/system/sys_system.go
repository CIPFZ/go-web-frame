package system

import (
	"github.com/CIPFZ/gowebframe/internal/config"
)

// System 配置文件结构体
type System struct {
	Config config.Config `json:"config"`
}
