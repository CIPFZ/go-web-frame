package router

import (
	"github.com/CIPFZ/gowebframe/internal/middleware"
	pluginApi "github.com/CIPFZ/gowebframe/internal/modules/plugin/api"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

type PluginRouter struct {
	svcCtx *svc.ServiceContext
	api    *pluginApi.PluginApi
}

func NewPluginRouter(svcCtx *svc.ServiceContext, api *pluginApi.PluginApi) *PluginRouter {
	return &PluginRouter{svcCtx: svcCtx, api: api}
}

func (r *PluginRouter) InitPluginRoutes(privateGroup *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	pluginGroup := privateGroup.Group("plugin")
	publicPluginGroup := publicGroup.Group("plugin/public")

	registryGroup := pluginGroup.Group("plugin")
	{
		registryGroup.POST("getPluginList", r.api.GetPluginList)
		registryGroup.GET("getPluginOverview", r.api.GetPluginOverview)
		registryGroup.POST("getProjectDetail", r.api.GetProjectDetail)
		registryWriteGroup := registryGroup.Group("", middleware.OperationRecord(r.svcCtx))
		{
			registryWriteGroup.POST("createPlugin", r.api.CreatePlugin)
			registryWriteGroup.PUT("updatePlugin", r.api.UpdatePlugin)
		}
	}

	releaseGroup := pluginGroup.Group("release")
	{
		releaseGroup.POST("getReleaseList", r.api.GetReleaseList)
		releaseGroup.POST("getReleaseDetail", r.api.GetReleaseDetail)
		releaseWriteGroup := releaseGroup.Group("", middleware.OperationRecord(r.svcCtx))
		{
			releaseWriteGroup.POST("createRelease", r.api.CreateRelease)
			releaseWriteGroup.PUT("updateRelease", r.api.UpdateRelease)
			releaseWriteGroup.POST("transition", r.api.TransitRelease)
			releaseWriteGroup.POST("assign", r.api.AssignRelease)
		}
	}

	{
		publicPluginGroup.POST("getPublishedPluginList", r.api.GetPublishedPluginList)
		publicPluginGroup.POST("getPublishedPluginDetail", r.api.GetPublishedPluginDetail)
	}
}
