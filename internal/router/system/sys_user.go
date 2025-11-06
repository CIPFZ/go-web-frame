package system

import (
	"github.com/CIPFZ/gowebframe/internal/api/v1/system"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
)

type UserRouter struct {
	serviceCtx *svc.ServiceContext
}

func (s *UserRouter) InitUserRouter(public, private *gin.RouterGroup) {
	userApi := system.NewUserApi(s.serviceCtx)

	// 1. 公共路由（无需鉴权）
	authGroup := public.Group("user")
	s.registerAuthRoutes(authGroup, userApi)

	// 2. 用户业务路由（带操作记录）
	userGroup := private.Group("user")
	s.registerUserRoutes(userGroup.Use(middleware.OperationRecord()), userApi)

	// 3. 用户只读接口（不带操作记录）
	s.registerReadOnlyUserRoutes(userGroup, userApi)
}

// 认证和登录类接口
func (s *UserRouter) registerAuthRoutes(group *gin.RouterGroup, api *system.UserApi) {
	group.POST("login", api.Login)
}

// 需要记录操作的用户接口
func (s *UserRouter) registerUserRoutes(group gin.IRoutes, api *system.UserApi) {
	group.POST("admin_register", api.Register)               // 管理员注册账号
	group.POST("changePassword", api.ChangePassword)         // 用户修改密码
	group.POST("setUserAuthority", api.SetUserAuthority)     // 设置用户权限
	group.DELETE("deleteUser", api.DeleteUser)               // 删除用户
	group.PUT("setUserInfo", api.SetUserInfo)                // 设置用户信息
	group.PUT("setSelfInfo", api.SetSelfInfo)                // 设置自身信息
	group.POST("setUserAuthorities", api.SetUserAuthorities) // 设置用户权限组
	group.POST("resetPassword", api.ResetPassword)           // 设置用户权限组
	group.PUT("setSelfSetting", api.SetSelfSetting)          // 用户界面配置
}

// 只读接口（不记录操作）
func (s *UserRouter) registerReadOnlyUserRoutes(group *gin.RouterGroup, api *system.UserApi) {
	group.POST("getUserList", api.GetUserList) // 分页获取用户列表
	group.POST("getUserInfo", api.GetUserInfo) // 获取自身信息
	group.GET("getSelfInfo", api.GetUserInfo)  // 获取自身信息
}
