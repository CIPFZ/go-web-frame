package router

import (
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/modules/agent/api"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

type Router struct {
	api    *api.AgentAPI
	svcCtx *svc.ServiceContext
}

func New(a *api.AgentAPI, ctx *svc.ServiceContext) *Router { return &Router{api: a, svcCtx: ctx} }
func (r *Router) Init(private *gin.RouterGroup) {
	group := private.Group("virtualization/vms/:name/agent")
	group.GET("", r.api.Status)
	group.GET("/health", r.api.Health)
	write := group.Group("", middleware.OperationRecord(r.svcCtx))
	write.POST("/tasks", r.api.CreateTask)
	private.GET("virtualization/agent/tasks/:taskID", r.api.Task)
	private.Group("", middleware.OperationRecord(r.svcCtx)).POST("virtualization/agent/tasks/:taskID/cancel", r.api.CancelTask)
}
