package api

import (
	"fmt"
	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/service"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// UserApi 定义了用户管理的 API
type UserApi struct {
	svcCtx      *svc.ServiceContext
	userService service.IUserService
}

// NewUserApi 创建一个新的 UserApi
func NewUserApi(svcCtx *svc.ServiceContext, userService service.IUserService) *UserApi {
	return &UserApi{
		svcCtx:      svcCtx,
		userService: userService,
	}
}

// Register 用户注册接口
// @Tags User
// @Summary 用户注册
// @Produce application/json
// @Param data body dto.RegisterReq true "注册参数"
// @Success 200 {object} response.Response{data=system.SysUser}
// @Router /user/register [post]
func (u *UserApi) Register(c *gin.Context) {
	var req dto.RegisterReq
	// 1. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数校验失败: "+err.Error(), c)
		return
	}

	// 2. 调用 Service (传入 Request Context 以保持链路追踪)
	user, err := u.userService.Register(c.Request.Context(), req)
	if err != nil {
		// Service 层返回的 error 已经是处理过的友好提示或记录过日志了
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 3. 返回结果 (注意隐藏密码)
	user.Password = ""
	response.OkWithData(user, c)
}

// Login 处理用户登录
// @Tags User
// @Summary 用户登录
// @Produce application/json
// @Param data body dto.LoginReq true "登录参数"
// @Success 200 {object} response.Response{data=dto.LoginResp}
// @Router /user/login [post]
func (u *UserApi) Login(c *gin.Context) {
	log := logger.GetLogger(c)
	var req dto.LoginReq
	err := c.ShouldBindJSON(&req)

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 验证码校验
	resp, err := u.userService.Login(c, req)
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

// setTokenHelper 在 cookie 和 header 中设置认证令牌
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

// SetSelfSetting 设置用户自己的配置
// @Tags User
// @Summary 设置用户配置
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body map[string]interface{} true "配置数据"
// @Success 200 {object} response.Response
// @Router /user/setSelfSetting [put]
func (u *UserApi) SetSelfSetting(c *gin.Context) {

}

// GetSelfInfo 获取当前用户的信息
// @Tags User
// @Summary 获取用户信息
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=system.SysUser}
// @Router /user/getSelfInfo [get]
func (u *UserApi) GetSelfInfo(c *gin.Context) {
	log := logger.GetLogger(c)
	// 1. 从 Context 中获取 Claims (由 JWTAuth 中间件注入)
	userUUID := utils.GetUserUUID(c)
	// 2. 调用 Service
	user, err := u.userService.GetUserInfo(c.Request.Context(), userUUID)
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
	err := u.userService.Logout(c.Request.Context(), token)
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
// @Tags User
// @Summary 获取用户列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body dto.SearchUserReq true "分页和搜索参数"
// @Success 200 {object} response.Response{data=dto.PageResult{list=[]system.SysUser}}
// @Router /user/getUserList [post]
func (u *UserApi) GetUserList(c *gin.Context) {
	var req dto.SearchUserReq
	_ = c.ShouldBindJSON(&req) // 允许空参
	log := logger.GetLogger(c)
	list, total, err := u.userService.GetUserList(c.Request.Context(), req)
	if err != nil {
		log.Error("get_user_list_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}

	// 确保密码字段不被返回 (Model中 json:"-" 已处理，这里是双保险)
	for i := range list {
		list[i].Password = ""
	}

	response.OkWithDetailed(common.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}

// AddUser 新增用户
// @Tags User
// @Summary 新增用户
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body dto.AddUserReq true "用户信息"
// @Success 200 {object} response.Response
// @Router /user/addUser [post]
func (u *UserApi) AddUser(c *gin.Context) {
	var req dto.AddUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.userService.AddUser(c.Request.Context(), req); err != nil {
		log.Error("add_user_error", zap.Error(err))
		response.FailWithMessage("添加失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// UpdateUser 更新已存在的用户
// @Tags User
// @Summary 更新用户
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body dto.UpdateUserReq true "用户信息"
// @Success 200 {object} response.Response
// @Router /user/updateUser [put]
func (u *UserApi) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.userService.UpdateUser(c.Request.Context(), req); err != nil {
		log.Error("update_user_error", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteUser 删除用户
// @Tags User
// @Summary 删除用户
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body dto.GetByIdReq true "用户 ID"
// @Success 200 {object} response.Response
// @Router /user/deleteUser [delete]
func (u *UserApi) DeleteUser(c *gin.Context) {
	var req common.GetByIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.userService.DeleteUser(c.Request.Context(), req.Uint()); err != nil {
		log.Error("delete_user_error", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ResetPassword 重置用户密码
// @Tags User
// @Summary 重置密码
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body dto.ResetPasswordReq true "用户 ID"
// @Success 200 {object} response.Response
// @Router /user/resetPassword [post]
func (u *UserApi) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := u.userService.ResetPassword(c.Request.Context(), req); err != nil {
		log.Error("reset_password_error", zap.Error(err))
		response.FailWithMessage("重置失败", c)
		return
	}
	response.OkWithMessage("重置成功", c)
}

// SwitchAuthority 切换用户当前的角色
// @Tags User
// @Summary 切换角色
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body Req true "角色 ID"
// @Success 200 {object} response.Response{data=dto.LoginResp}
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
	resp, err := u.userService.SwitchAuthority(c.Request.Context(), userUUID, r.AuthorityId)
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

// UpdateSelfInfo 更新个人信息 (昵称、简介)
// @Router /user/info [put]
func (u *UserApi) UpdateSelfInfo(c *gin.Context) {
	var req dto.UpdateSelfInfoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	uid := utils.GetUserUUID(c)
	err := u.userService.UpdateSelfInfo(c.Request.Context(), uid, req)
	if err != nil {
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// UpdateUiConfig 更新 UI 配置
// @Router /user/ui-config [put]
func (u *UserApi) UpdateUiConfig(c *gin.Context) {
	var req dto.UpdateUiConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("配置参数错误", c)
		return
	}
	uid := utils.GetUserUUID(c)

	if err := u.userService.UpdateUiConfig(c.Request.Context(), uid, req); err != nil {
		response.FailWithMessage("配置保存失败", c)
		return
	}
	response.OkWithMessage("保存", c)
}

// UploadAvatar 上传并更新头像
func (u *UserApi) UploadAvatar(c *gin.Context) {
	log := logger.GetLogger(c)
	// 1. 获取文件
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择要上传的图片", c)
		return
	}

	// 2. 获取当前用户ID
	uid := utils.GetUserUUID(c)
	fileName := fmt.Sprintf("user/avatar/%s.png", uid)
	// 3. ✨ 调用 FileService 进行物理上传
	// 这步会自动处理：检查后缀、大小、上传到MinIO/Local、写入sys_files表、链路追踪
	avatarUrl, _, err := u.svcCtx.OSS.Upload(c.Request.Context(), header, fileName)
	if err != nil {
		log.Error("avatar_upload_failed", zap.Error(err))
		response.FailWithMessage("头像上传失败: "+err.Error(), c)
		return
	}

	// 4. ✨ 调用 UserService 更新用户表的 avatar 字段
	if err := u.userService.UpdateAvatar(c.Request.Context(), uid, avatarUrl); err != nil {
		log.Error("avatar_update_db_failed", zap.Error(err))
		// 这是一个边缘情况：文件传上去了但资料没改掉。
		// 实际上为了严谨可以考虑异步删除刚才的文件，但头像场景通常忽略回滚
		response.FailWithMessage("更新用户资料失败", c)
		return
	}

	// 5. 返回新头像 URL
	response.OkWithData(gin.H{"url": avatarUrl}, c)
}
