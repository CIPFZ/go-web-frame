package system

import (
	"github.com/CIPFZ/gowebframe/internal/api/v1/system"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
)

type ApiRouter struct {
	serviceCtx *svc.ServiceContext
}

func (a *ApiRouter) InitApiRouter(privateGroup *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	apiRouterApi := system.NewSysApiApi(a.serviceCtx)
	// 带操作记录中间件的
	apiPrivateWithRecord := privateGroup.Group("api")
	apiPrivateWithRecord.Use(middleware.OperationRecord(a.serviceCtx))
	a.registerPrivateWithRecord(apiPrivateWithRecord, apiRouterApi)

	// 不带中间件
	apiPrivate := privateGroup.Group("api")
	a.registerPrivate(apiPrivate, apiRouterApi)

	// 公共路由
	apiPublic := publicGroup.Group("api")
	a.registerPublic(apiPublic, apiRouterApi)
}

func (a *ApiRouter) registerPrivate(router *gin.RouterGroup, apiRouterApi *system.SysApiApi) {
	//router.POST("getAllApis", apiRouterApi.GetAllApis)
	router.POST("getApiList", apiRouterApi.GetApiList)
}

func (a *ApiRouter) registerPrivateWithRecord(router *gin.RouterGroup, apiRouterApi *system.SysApiApi) {
	//router.GET("getApiGroups", apiRouterApi.GetApiGroups)
	//router.GET("syncApi", apiRouterApi.SyncApi)
	//router.POST("ignoreApi", apiRouterApi.IgnoreApi)
	//router.POST("enterSyncApi", apiRouterApi.EnterSyncApi)
	router.POST("createApi", apiRouterApi.CreateApi)
	router.POST("deleteApi", apiRouterApi.DeleteApi)
	//router.POST("getApiById", apiRouterApi.GetApiById)
	router.POST("updateApi", apiRouterApi.UpdateApi)
	//router.DELETE("deleteApisByIds", apiRouterApi.DeleteApisByIds)
}

func (a *ApiRouter) registerPublic(router *gin.RouterGroup, apiRouterApi *system.SysApiApi) {
	//router.GET("freshCasbin", apiRouterApi.FreshCasbin)
}
