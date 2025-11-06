package system

import (
	"github.com/CIPFZ/gowebframe/internal/api/v1/system"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
)

type CasbinRouter struct {
	serviceCtx *svc.ServiceContext
}

func (s *CasbinRouter) InitCasbinRouter(Router *gin.RouterGroup) {
	casbinApi := system.NewCasbinApi(s.serviceCtx)
	casbinRouter := Router.Group("casbin").Use(middleware.OperationRecord())
	casbinRouter.POST("updateCasbin", casbinApi.UpdateCasbin)

	casbinRouterWithoutRecord := Router.Group("casbin")
	casbinRouterWithoutRecord.POST("getPolicyPathByAuthorityId", casbinApi.GetPolicyPathByAuthorityId)
}
