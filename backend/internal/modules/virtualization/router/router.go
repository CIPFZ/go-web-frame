package router

import (
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/modules/virtualization/api"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

type Router struct {
	api     *api.VirtualMachineAPI
	storage *api.StorageAPI
	svcCtx  *svc.ServiceContext
}

func New(api *api.VirtualMachineAPI, storage *api.StorageAPI, svcCtx *svc.ServiceContext) *Router {
	return &Router{api: api, storage: storage, svcCtx: svcCtx}
}

func (r *Router) Init(private *gin.RouterGroup) {
	group := private.Group("virtualization/vms")
	group.GET("", r.api.List)
	group.GET("/:name", r.api.Detail)

	write := group.Group("", middleware.OperationRecord(r.svcCtx))
	write.POST("", r.api.Create)
	write.PUT("/:name", r.api.Update)
	write.POST("/:name/start", r.api.Start)
	write.POST("/:name/shutdown", r.api.Shutdown)
	write.POST("/:name/reboot", r.api.Reboot)
	write.POST("/:name/force-stop", r.api.ForceStop)
	write.DELETE("/:name", r.api.Delete)

	storage := private.Group("virtualization/storage")
	storage.GET("/pools", r.storage.ListPools)
	storage.GET("/volumes", r.storage.ListVolumes)
	storageWrite := storage.Group("", middleware.OperationRecord(r.svcCtx))
	storageWrite.POST("/volumes", r.storage.CreateVolume)
	storageWrite.PUT("/volumes/:pool/:name", r.storage.ResizeVolume)
	storageWrite.DELETE("/volumes/:pool/:name", r.storage.DeleteVolume)
}
