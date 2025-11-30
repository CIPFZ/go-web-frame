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
	s.registerUserRoutes(userGroup.Use(middleware.OperationRecord(s.serviceCtx)), userApi)

	// 3. 用户只读接口（不带操作记录）
	s.registerReadOnlyUserRoutes(userGroup, userApi)
}

// 认证和登录类接口
func (s *UserRouter) registerAuthRoutes(group *gin.RouterGroup, api *system.UserApi) {
	group.POST("login", api.Login)
}

// 需要记录操作的用户接口
func (s *UserRouter) registerUserRoutes(group gin.IRoutes, api *system.UserApi) {
	group.POST("registerUser", api.Register)        // 管理员注册账号
	group.PUT("setSelfSetting", api.SetSelfSetting) // 用户界面配置
	group.POST("updateUser", api.UpdateUser)        // 设置用户信息
	group.POST("registerAdmin", api.AddUser)        // 管理员创建用户
	group.POST("deleteUser", api.DeleteUser)        // 用户删除
}

// 只读接口（不记录操作）
func (s *UserRouter) registerReadOnlyUserRoutes(group *gin.RouterGroup, api *system.UserApi) {
	group.POST("getUserList", api.GetUserList) // 分页获取用户列表
	//group.POST("getUserInfo", api.GetUserInfo)         // 获取用户信息
	group.GET("getSelfInfo", api.GetSelfInfo)          // 获取自身信息
	group.GET("logout", api.Logout)                    // 退出登录
	group.POST("resetPassword", api.ResetPassword)     // 更新密码
	group.POST("switchAuthority", api.SwitchAuthority) // 用户切换角色
}
