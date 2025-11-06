package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthorityMenuApi struct {
	svcCtx          *svc.ServiceContext
	menuService     systemService.IMenuService
	baseMenuService systemService.IBaseMenuService
}

func NewAuthorityMenuApi(svcCtx *svc.ServiceContext) *AuthorityMenuApi {
	return &AuthorityMenuApi{
		svcCtx:          svcCtx,
		menuService:     systemService.NewMenuService(svcCtx),
		baseMenuService: systemService.NewBaseMenuService(svcCtx),
	}
}

// GetMenu 获取用户动态路由
func (a *AuthorityMenuApi) GetMenu(c *gin.Context) {
	log := logger.GetLogger(c)
	menus, err := a.menuService.GetMenuTree(utils.GetUserAuthorityId(c, a.svcCtx))
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	if menus == nil {
		menus = []systemModel.SysBaseMenu{}
	}
	response.OkWithDetailed(menus, "获取成功", c)
}

// GetBaseMenuTree 获取用户动态路由
func (a *AuthorityMenuApi) GetBaseMenuTree(c *gin.Context) {
	log := logger.GetLogger(c)
	authority := utils.GetUserAuthorityId(c, a.svcCtx)
	menus, err := a.menuService.GetBaseMenuTree(authority)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(systemRes.SysBaseMenusResponse{Menus: menus}, "获取成功", c)
}

// AddMenuAuthority 增加menu和角色关联关系
func (a *AuthorityMenuApi) AddMenuAuthority(c *gin.Context) {
	var authorityMenu systemReq.AddMenuAuthorityInfo
	err := c.ShouldBindJSON(&authorityMenu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	adminAuthorityID := utils.GetUserAuthorityId(c, a.svcCtx)
	if err := a.menuService.AddMenuAuthority(authorityMenu.Menus, adminAuthorityID, authorityMenu.AuthorityId); err != nil {
		log.Error("添加失败!", zap.Error(err))
		response.FailWithMessage("添加失败", c)
	} else {
		response.OkWithMessage("添加成功", c)
	}
}

// GetMenuAuthority 获取指定角色menu
func (a *AuthorityMenuApi) GetMenuAuthority(c *gin.Context) {
	var param request.GetAuthorityId
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	menus, err := a.menuService.GetMenuAuthority(&param)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithDetailed(systemRes.SysMenusResponse{Menus: menus}, "获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"menus": menus}, "获取成功", c)
}

// AddBaseMenu 新增菜单
func (a *AuthorityMenuApi) AddBaseMenu(c *gin.Context) {
	var menu systemModel.SysBaseMenu
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	err = a.menuService.AddBaseMenu(menu)
	if err != nil {
		log.Error("添加失败!", zap.Error(err))
		response.FailWithMessage("添加失败："+err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// DeleteBaseMenu 删除菜单
func (a *AuthorityMenuApi) DeleteBaseMenu(c *gin.Context) {
	var menu request.GetById
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	err = a.baseMenuService.DeleteBaseMenu(c, menu.ID)
	if err != nil {
		log.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateBaseMenu 更新菜单
func (a *AuthorityMenuApi) UpdateBaseMenu(c *gin.Context) {
	var menu systemModel.SysBaseMenu
	err := c.ShouldBindJSON(&menu)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	err = a.baseMenuService.UpdateBaseMenu(c, menu)
	if err != nil {
		log.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetBaseMenuById 根据id获取菜单
func (a *AuthorityMenuApi) GetBaseMenuById(c *gin.Context) {
	var idInfo request.GetById
	err := c.ShouldBindJSON(&idInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	menu, err := a.baseMenuService.GetBaseMenuById(c, idInfo.ID)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(systemRes.SysBaseMenuResponse{Menu: menu}, "获取成功", c)
}

// GetMenuList 页获取基础menu列表
func (a *AuthorityMenuApi) GetMenuList(c *gin.Context) {
	authorityID := utils.GetUserAuthorityId(c, a.svcCtx)
	log := logger.GetLogger(c)
	menuList, err := a.menuService.GetInfoList(authorityID)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(menuList, "获取成功", c)
}
