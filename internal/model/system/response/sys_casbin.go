package response

import "github.com/CIPFZ/gowebframe/internal/model/system/request"

type PolicyPathResponse struct {
	Paths []request.CasbinInfo `json:"paths"`
}
