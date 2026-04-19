package router

import (
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/api"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

type PluginRouter struct {
	svcCtx *svc.ServiceContext
	api    *api.PluginApi
}

func NewPluginRouter(svcCtx *svc.ServiceContext, pluginAPI *api.PluginApi) *PluginRouter {
	return &PluginRouter{svcCtx: svcCtx, api: pluginAPI}
}

func (r *PluginRouter) InitPluginRoutes(privateGroup, publicGroup *gin.RouterGroup) {
	public := publicGroup.Group("plugin")
	{
		publicAPI := public.Group("public")
		{
			publicAPI.POST("getPublishedPluginList", r.api.GetPublishedPluginList)
			publicAPI.POST("getPublishedPluginDetail", r.api.GetPublishedPluginDetail)
		}
	}

	{
		pluginGroup := privateGroup.Group("plugin/plugin")
		{
			pluginGroup.POST("getPluginList", r.api.GetPluginList)
			pluginGroup.POST("getProjectDetail", r.api.GetProjectDetail)

			write := pluginGroup.Group("", middleware.OperationRecord(r.svcCtx))
			write.POST("createPlugin", r.api.CreatePlugin)
			write.PUT("updatePlugin", r.api.UpdatePlugin)
		}

		releaseGroup := privateGroup.Group("plugin/release")
		{
			releaseGroup.POST("getReleaseDetail", r.api.GetReleaseDetail)
			write := releaseGroup.Group("", middleware.OperationRecord(r.svcCtx))
			write.POST("createRelease", r.api.CreateRelease)
			write.PUT("updateRelease", r.api.UpdateRelease)
			write.POST("transition", r.api.TransitionRelease)
			write.POST("claim", r.api.ClaimWorkOrder)
			write.POST("reset", r.api.ResetWorkOrder)
		}

		workOrderGroup := privateGroup.Group("plugin/work-order")
		workOrderGroup.POST("getWorkOrderPool", r.api.GetWorkOrderPool)

		productGroup := privateGroup.Group("plugin/product")
		{
			productGroup.POST("getProductList", r.api.GetProductList)
			write := productGroup.Group("", middleware.OperationRecord(r.svcCtx))
			write.POST("createProduct", r.api.CreateProduct)
			write.PUT("updateProduct", r.api.UpdateProduct)
		}

		departmentGroup := privateGroup.Group("plugin/department")
		{
			departmentGroup.POST("getDepartmentList", r.api.GetDepartmentList)
			write := departmentGroup.Group("", middleware.OperationRecord(r.svcCtx))
			write.POST("createDepartment", r.api.CreateDepartment)
			write.PUT("updateDepartment", r.api.UpdateDepartment)
		}
	}
}
