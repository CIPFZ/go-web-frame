package browserproxy

import (
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service: service} }

func (a *API) Bootstrap(c *gin.Context) {
	response.OkWithData(a.service.Bootstrap(c.Request.Context()), c)
}

func (a *API) Health(c *gin.Context) {
	response.OkWithData(a.service.Health(c.Request.Context()), c)
}
