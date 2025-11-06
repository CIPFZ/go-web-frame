package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthorityApi struct {
	svcCtx  *svc.ServiceContext
	service systemService.IAuthorityService
}

func NewAuthorityApi(svcCtx *svc.ServiceContext) *AuthorityApi {
	return &AuthorityApi{
		svcCtx:  svcCtx,
		service: systemService.NewAuthorityService(svcCtx),
	}
}

// CreateAuthority 创建角色
func (a *AuthorityApi) CreateAuthority(c *gin.Context) {
	var authority, authBack systemModel.SysAuthority
	var err error

	if err = c.ShouldBindJSON(&authority); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// TODO 参数校验
	log := logger.GetLogger(c)
	if *authority.ParentId == 0 && a.svcCtx.Config.System.UseStrictAuth {
		authority.ParentId = utils.Pointer(utils.GetUserAuthorityId(c, a.svcCtx))
	}

	if authBack, err = a.service.CreateAuthority(authority); err != nil {
		log.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败"+err.Error(), c)
		return
	}
	err = systemService.NewCasbinService(a.svcCtx).FreshCasbin()
	if err != nil {
		log.Error("创建成功，权限刷新失败。", zap.Error(err))
		response.FailWithMessage("创建成功，权限刷新失败。"+err.Error(), c)
		return
	}
	response.OkWithDetailed(systemRes.SysAuthorityResponse{Authority: authBack}, "创建成功", c)
}

// CopyAuthority 拷贝角色
func (a *AuthorityApi) CopyAuthority(c *gin.Context) {
	var copyInfo systemRes.SysAuthorityCopyResponse
	err := c.ShouldBindJSON(&copyInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	adminAuthorityID := utils.GetUserAuthorityId(c, a.svcCtx)
	authBack, err := a.service.CopyAuthority(adminAuthorityID, copyInfo)
	if err != nil {
		log.Error("拷贝失败!", zap.Error(err))
		response.FailWithMessage("拷贝失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(systemRes.SysAuthorityResponse{Authority: authBack}, "拷贝成功", c)
}

// DeleteAuthority 删除角色
func (a *AuthorityApi) DeleteAuthority(c *gin.Context) {
	var authority systemModel.SysAuthority
	var err error
	if err = c.ShouldBindJSON(&authority); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	// 删除角色之前需要判断是否有用户正在使用此角色
	if err = a.service.DeleteAuthority(&authority); err != nil {
		log.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败"+err.Error(), c)
		return
	}
	_ = systemService.NewCasbinService(a.svcCtx).FreshCasbin()
	response.OkWithMessage("删除成功", c)
}

// UpdateAuthority 更新角色信息
func (a *AuthorityApi) UpdateAuthority(c *gin.Context) {
	var auth systemModel.SysAuthority
	err := c.ShouldBindJSON(&auth)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	authority, err := a.service.UpdateAuthority(c, auth)
	if err != nil {
		log.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(systemRes.SysAuthorityResponse{Authority: authority}, "更新成功", c)
}

// GetAuthorityList 分页获取角色列表
func (a *AuthorityApi) GetAuthorityList(c *gin.Context) {
	log := logger.GetLogger(c)
	authorityID := utils.GetUserAuthorityId(c, a.svcCtx)
	list, err := a.service.GetAuthorityInfoList(authorityID)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SetDataAuthority 设置角色资源权限
func (a *AuthorityApi) SetDataAuthority(c *gin.Context) {
	var auth systemModel.SysAuthority
	err := c.ShouldBindJSON(&auth)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	adminAuthorityID := utils.GetUserAuthorityId(c, a.svcCtx)
	err = a.service.SetDataAuthority(adminAuthorityID, auth)
	if err != nil {
		log.Error("设置失败!", zap.Error(err))
		response.FailWithMessage("设置失败"+err.Error(), c)
		return
	}
	response.OkWithMessage("设置成功", c)
}
