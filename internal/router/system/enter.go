package system

import (
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

type Group struct {
	serviceCtx            *svc.ServiceContext
	ApiRouter             ApiRouter             // 接口路由
	AuthorityRouter       AuthorityRouter       // 权限路由
	AuthorityBtnRouter    AuthorityBtnRouter    // 权限按钮路由
	CasbinRouter          CasbinRouter          // 访问控制路由
	MenuRouter            MenuRouter            // 菜单路由
	OperationRecordRouter OperationRecordRouter // 操作记录路由
	UserRouter            UserRouter            // 用户路由
	SysRouter             SysRouter             // 系统信息路由
}

// NewSystemGroup 是构造函数，替代原来的全局 var
func NewSystemGroup(serviceCtx *svc.ServiceContext) *Group {
	return &Group{
		serviceCtx:            serviceCtx,
		ApiRouter:             ApiRouter{serviceCtx: serviceCtx},
		AuthorityRouter:       AuthorityRouter{serviceCtx: serviceCtx},
		AuthorityBtnRouter:    AuthorityBtnRouter{serviceCtx: serviceCtx},
		CasbinRouter:          CasbinRouter{serviceCtx: serviceCtx},
		MenuRouter:            MenuRouter{serviceCtx: serviceCtx},
		OperationRecordRouter: OperationRecordRouter{serviceCtx: serviceCtx},
		UserRouter:            UserRouter{serviceCtx: serviceCtx},
		SysRouter:             SysRouter{serviceCtx: serviceCtx},
	}
}

// RegisterRoutes 统一注册 system 模块下的所有子路由
func (g *Group) RegisterRoutes(privateGroup *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	// 公共路由（不需要鉴权）
	// 私有路由（需要鉴权）
	g.ApiRouter.InitApiRouter(privateGroup, publicGroup)
	g.AuthorityRouter.InitAuthorityRouter(privateGroup)
	g.AuthorityBtnRouter.InitAuthorityBtnRouterRouter(privateGroup)
	g.CasbinRouter.InitCasbinRouter(privateGroup)
	g.MenuRouter.InitMenuRouter(privateGroup)
	g.OperationRecordRouter.InitSysOperationRecordRouter(privateGroup)
	g.UserRouter.InitUserRouter(publicGroup, privateGroup)
	g.SysRouter.InitSystemRouter(privateGroup)
}
