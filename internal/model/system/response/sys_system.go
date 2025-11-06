package response

import (
	"github.com/CIPFZ/gowebframe/internal/config"
)

type SysConfigResponse struct {
	Config *config.Config `json:"config"`
}
