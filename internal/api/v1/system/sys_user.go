package system

import (
	"time"

	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
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

func NewUserApi(serviceCtx *svc.ServiceContext) *UserApi {
	return &UserApi{
		svcCtx:  serviceCtx,
		service: systemService.NewUserService(serviceCtx),
	}
}

// Login 登录
func (u *UserApi) Login(c *gin.Context) {
	log := logger.GetLogger(c)
	var l systemReq.Login
	err := c.ShouldBindJSON(&l)

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	model := &systemModel.SysUser{Username: l.Username, Password: l.Password}
	user, err := u.service.Login(c, model)
	if err != nil {
		log.Error("登陆失败! 用户名不存在或者密码错误!", zap.Error(err))
		response.FailWithMessage("用户名不存在或者密码错误", c)
		return
	}
	if user.Enable != 1 {
		log.Error("登陆失败! 用户被禁止登录!")
		response.FailWithMessage("用户被禁止登录", c)
		return
	}
	u.TokenNext(c, *user)
}

// TokenNext 登录以后签发jwt
func (u *UserApi) TokenNext(c *gin.Context, user systemModel.SysUser) {
	log := logger.GetLogger(c)
	token, claims, err := u.LoginToken(&user)
	if err != nil {
		log.Error("获取token失败!", zap.Error(err))
		response.FailWithMessage("获取token失败", c)
		return
	}
	utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
	response.OkWithDetailed(systemRes.LoginResponse{
		Token:     token,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, "登录成功", c)

	// TODO 单点登录
}

// LoginToken 用户登录token
func (u *UserApi) LoginToken(user systemModel.Login) (string, systemReq.CustomClaims, error) {
	var token string
	var claims systemReq.CustomClaims
	var err error
	j := utils.NewJWT(u.svcCtx)
	claims = j.CreateClaims(systemReq.BaseClaims{
		UUID:        user.GetUUID(),
		ID:          user.GetUserId(),
		NickName:    user.GetNickname(),
		Username:    user.GetUsername(),
		AuthorityId: user.GetAuthorityId(),
	})
	token, err = j.CreateToken(claims)
	return token, claims, err
}

// Register 注册
func (u *UserApi) Register(c *gin.Context) {
}

// ChangePassword 用户修改密码
func (u *UserApi) ChangePassword(c *gin.Context) {
}

// GetUserList 分页获取用户列表
func (u *UserApi) GetUserList(c *gin.Context) {
	log := logger.GetLogger(c)
	var pageInfo systemReq.GetUserList
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验

	list, total, err := u.service.GetUserInfoList(c, pageInfo)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)

}

// SetUserAuthority 更改用户权限
func (u *UserApi) SetUserAuthority(c *gin.Context) {
	var sua systemReq.SetUserAuth
	err := c.ShouldBindJSON(&sua)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	//log := logger.GetLogger(c)
	//userID := utils.GetUserID(c)
	//err = u.service.SetUserAuthority(userID, sua.AuthorityId)
	//if err != nil {
	//	log.Error("修改失败!", zap.Error(err))
	//	response.FailWithMessage(err.Error(), c)
	//	return
	//}
	//claims := utils.GetUserInfo(c)
	//claims.AuthorityId = sua.AuthorityId
	//token, err := utils.NewJWT(u.svcCtx).CreateToken(*claims)
	//if err != nil {
	//	log.Error("修改失败!", zap.Error(err))
	//	response.FailWithMessage(err.Error(), c)
	//	return
	//}
	//c.Header("new-token", token)
	//c.Header("new-expires-at", strconv.FormatInt(claims.ExpiresAt.Unix(), 10))
	//utils.SetToken(c, token, int((claims.ExpiresAt.Unix()-time.Now().Unix())/60))
	response.OkWithMessage("修改成功", c)
}

// SetUserAuthorities 设置用户权限
func (u *UserApi) SetUserAuthorities(c *gin.Context) {
	//var sua systemReq.SetUserAuthorities
	//err := c.ShouldBindJSON(&sua)
	//if err != nil {
	//	response.FailWithMessage(err.Error(), c)
	//	return
	//}
	//authorityID := utils.GetUserAuthorityId(c)
	//err = userService.SetUserAuthorities(authorityID, sua.ID, sua.AuthorityIds)
	//if err != nil {
	//	global.GVA_LOG.Error("修改失败!", zap.Error(err))
	//	response.FailWithMessage("修改失败", c)
	//	return
	//}
	response.OkWithMessage("修改成功", c)
}

// DeleteUser 删除用户
func (u *UserApi) DeleteUser(c *gin.Context) {

}

// SetUserInfo 设置用户信息
func (u *UserApi) SetUserInfo(c *gin.Context) {

}

// SetSelfInfo 设置用户信息
func (u *UserApi) SetSelfInfo(c *gin.Context) {

}

// SetSelfSetting 设置用户配置
func (u *UserApi) SetSelfSetting(c *gin.Context) {

}

// GetUserInfo 获取用户信息
func (u *UserApi) GetUserInfo(c *gin.Context) {
	log := logger.GetLogger(c)
	// 从 token 中获取用户ID
	uuid := utils.GetUserUuid(c, u.svcCtx)
	// 根据用户ID获取用户信息
	userInfo, err := u.service.GetUserInfo(c, uuid)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(userInfo, "获取成功", c)
}

// ResetPassword 重置用户密码
func (u *UserApi) ResetPassword(c *gin.Context) {

}
