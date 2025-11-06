package system

import (
	"github.com/CIPFZ/gowebframe/internal/api/v1/system"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
)

type AuthorityRouter struct {
	serviceCtx *svc.ServiceContext
}

func (s *AuthorityRouter) InitAuthorityRouter(Router *gin.RouterGroup) {
	authorityRouter := Router.Group("authority").Use(middleware.OperationRecord())
	authorityApi := system.NewAuthorityApi(s.serviceCtx)
	authorityRouter.POST("createAuthority", authorityApi.CreateAuthority)   // 创建角色
	authorityRouter.POST("deleteAuthority", authorityApi.DeleteAuthority)   // 删除角色
	authorityRouter.PUT("updateAuthority", authorityApi.UpdateAuthority)    // 更新角色
	authorityRouter.POST("copyAuthority", authorityApi.CopyAuthority)       // 拷贝角色
	authorityRouter.POST("setDataAuthority", authorityApi.SetDataAuthority) // 设置角色资源权限

	authorityRouterWithoutRecord := Router.Group("authority")
	authorityRouterWithoutRecord.POST("getAuthorityList", authorityApi.GetAuthorityList) // 获取角色列表
}
