package system

import (
	"net/http"
	"time"

	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type UserApi struct {
	svcCtx  *svc.ServiceContext
	service systemService.IUserService
}

func NewUserApi(svcCtx *svc.ServiceContext) *UserApi {
	return &UserApi{
		svcCtx:  svcCtx,
		service: systemService.NewUserService(svcCtx),
	}
}

// Register 用户注册接口
// @Tags User
// @Summary 用户注册
// @Produce application/json
// @Param data body system.RegisterReq true "注册参数"
// @Success 200 {object} response.Response{data=system.SysUser}
// @Router /user/register [post]
func (u *UserApi) Register(c *gin.Context) {
	var req systemReq.RegisterReq
	// 1. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数校验失败: "+err.Error(), c)
		return
	}

	// 2. 调用 Service (传入 Request Context 以保持链路追踪)
	user, err := u.service.Register(c.Request.Context(), req)
	if err != nil {
		// Service 层返回的 error 已经是处理过的友好提示或记录过日志了
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 3. 返回结果 (注意隐藏密码)
	user.Password = ""
	response.OkWithData(user, c)
}

// Login 登录
func (u *UserApi) Login(c *gin.Context) {
	log := logger.GetLogger(c)
	var req systemReq.Login
	err := c.ShouldBindJSON(&req)

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 验证码校验
	resp, err := u.service.Login(c, req)
	if err != nil {
		log.Error("登陆失败! 用户名不存在或者密码错误!", zap.Error(err))
		response.FailWithMessage("用户名不存在或者密码错误", c)
		return
	}

	now := time.Now().Unix()
	maxAge := int((resp.ExpiresAt / 1000) - now)

	u.setTokenHelper(c, resp.Token, maxAge)

	// 3. 返回 JSON
	response.OkWithDetailed(resp, "登录成功", c)
}

// setTokenHelper 设置 Cookie
func (u *UserApi) setTokenHelper(c *gin.Context, token string, maxAge int) {
	// 2. 智能判断 Domain
	// 本地调试 (localhost, 127.0.0.1) -> 留空，兼容性最好
	// 生产环境 -> 也可以留空，浏览器会自动将其限制在当前域名下
	// 除非你需要跨子域共享 (如 a.site.com 和 b.site.com 共享)，才需要配置为 ".site.com"
	var domain string

	// 如果你有明确的配置需求，可以从 Config 读取，否则默认为空
	if u.svcCtx.Config.System.Environment == "production" {
		domain = ""
	}

	// 3. 智能判断 Secure
	// 生产环境通常使用 HTTPS，所以 Secure = true
	isSecure := u.svcCtx.Config.System.Environment == "production"

	// 4. 设置 Cookie
	c.SetCookie(
		"x-token", // name
		token,     // value
		maxAge,    // maxAge (秒)
		"/",       // path
		domain,    // domain (留空是最稳健的选择)
		isSecure,  // secure (HTTPS only)
		true,      // httpOnly (关键安全设置: 禁止 JS 读取)
	)

	// 5. 设置 SameSite (防止 CSRF)
	c.SetSameSite(http.SameSiteLaxMode)

	// 6. 同时将 Token 放入 Header (方便前端拦截器抓取)
	c.Header("x-token", token)
	c.Header("new-token", token)
}

// SetSelfSetting 设置用户配置
func (u *UserApi) SetSelfSetting(c *gin.Context) {

}

// GetSelfInfo 用户获取自己信息
func (u *UserApi) GetSelfInfo(c *gin.Context) {
	log := logger.GetLogger(c)
	// 1. 从 Context 中获取 Claims (由 JWTAuth 中间件注入)
	userUUID := utils.GetUserUUID(c)
	// 2. 调用 Service
	user, err := u.service.GetUserInfo(c.Request.Context(), userUUID)
	if err != nil {
		log.Error("get_user_info_failed", zap.Error(err))
		response.FailWithMessage("获取用户信息失败", c)
		return
	}

	response.OkWithData(user, c)
}

// Logout 用户登出
// @Tags User
// @Summary 用户登出
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{msg=string}
// @Router /user/logout [post]
func (u *UserApi) Logout(c *gin.Context) {
	// 1. 获取 Token (从 Header 中)
	token := c.GetHeader("x-token")
	if token == "" {
		// 如果没有 Token，视为已登出
		response.OkWithMessage("注销成功", c)
		return
	}
	log := logger.GetLogger(c)
	// 2. 调用 Service
	err := u.service.Logout(c.Request.Context(), token)
	if err != nil {
		// 即使 Redis 写入失败，通常也应该告诉前端“登出成功”，
		// 并在后端记录错误日志（Service 层已做），避免阻塞用户退出 UI。
		log.Error("logout_api_error", zap.Error(err))
		response.FailWithMessage("注销失败，请重试", c)
		return
	}

	// 3. 响应
	response.OkWithMessage("注销成功", c)
}

// GetUserList 分页获取用户列表
// @Router /user/getUserList [post]
func (u *UserApi) GetUserList(c *gin.Context) {
	var req systemReq.SearchUserReq
	_ = c.ShouldBindJSON(&req) // 允许空参
	log := logger.GetLogger(c)
	list, total, err := u.service.GetUserList(c.Request.Context(), req)
	if err != nil {
		log.Error("get_user_list_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}

	// 确保密码字段不被返回 (Model中 json:"-" 已处理，这里是双保险)
	for i := range list {
		list[i].Password = ""
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}

// AddUser 新增用户
func (u *UserApi) AddUser(c *gin.Context) {
	var req systemReq.AddUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.service.AddUser(c.Request.Context(), req); err != nil {
		log.Error("add_user_error", zap.Error(err))
		response.FailWithMessage("添加失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// UpdateUser 更新用户
func (u *UserApi) UpdateUser(c *gin.Context) {
	var req systemReq.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.service.UpdateUser(c.Request.Context(), req); err != nil {
		log.Error("update_user_error", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteUser 删除用户
func (u *UserApi) DeleteUser(c *gin.Context) {
	var req request.GetByIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.service.DeleteUser(c.Request.Context(), req.Uint()); err != nil {
		log.Error("delete_user_error", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ResetPassword 重置密码
func (u *UserApi) ResetPassword(c *gin.Context) {
	var req systemReq.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.service.ResetPassword(c.Request.Context(), req); err != nil {
		log.Error("reset_password_error", zap.Error(err))
		response.FailWithMessage("重置失败", c)
		return
	}
	response.OkWithMessage("重置成功", c)
}

// SwitchAuthority 切换当前角色
// @Router /user/switchAuthority [post]
func (u *UserApi) SwitchAuthority(c *gin.Context) {
	type Req struct {
		AuthorityId uint `json:"authorityId" binding:"required"`
	}
	var r Req
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userUUID := utils.GetUserUUID(c)

	// 调用 Service
	resp, err := u.service.SwitchAuthority(c.Request.Context(), userUUID, r.AuthorityId)
	if err != nil {
		response.FailWithMessage("切换失败: "+err.Error(), c)
		return
	}

	// 更新 Cookie 和 Header
	now := time.Now().Unix()
	maxAge := int((resp.ExpiresAt / 1000) - now)
	u.setTokenHelper(c, resp.Token, maxAge)
	c.Header("new-token", resp.Token)

	response.OkWithData(resp, c)
}
