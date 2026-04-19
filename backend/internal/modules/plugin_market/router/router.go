package router

import (
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/api"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine, apiHandler *api.Api) {
	v1 := r.Group("/api/v1/market")
	{
		v1.POST("/plugins/list", apiHandler.ListPlugins)
		v1.POST("/plugins/detail", apiHandler.GetPluginDetail)
		sync := v1.Group("/sync")
		{
			sync.POST("/plugin/upsert", apiHandler.UpsertPlugin)
			sync.POST("/plugin/delete", apiHandler.DeletePlugin)
			sync.POST("/version/upsert", apiHandler.UpsertVersion)
			sync.POST("/version/offline", apiHandler.OfflineVersion)
		}
	}
}
