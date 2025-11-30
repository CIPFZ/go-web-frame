package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MenuApi struct {
	svcCtx      *svc.ServiceContext
	menuService systemService.IMenuService
}

func NewMenuApi(svcCtx *svc.ServiceContext) *MenuApi {
	return &MenuApi{
		svcCtx:      svcCtx,
		menuService: systemService.NewMenuService(svcCtx),
	}
}

// GetMenu 获取当前用户的动态菜单
// @Tags Menu
// @Summary 获取动态菜单
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]system.SysMenu}
// @Router /menu/getMenu [get]
func (a *MenuApi) GetMenu(c *gin.Context) {
	// 1. 从 Context 获取当前用户的角色 ID
	authorityId := utils.GetAuthorityId(c)
	if authorityId == 0 {
		response.FailWithMessage("无法获取用户角色信息", c)
		return
	}

	// 2. 调用 Service
	menus, err := a.menuService.GetUserMenuTree(c.Request.Context(), authorityId)
	if err != nil {
		// 错误日志已在 Service 层记录
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 3. 返回结果
	// 注意：这里返回的是 []SysMenu，前端(layout.menu.request)会接收到这个数组
	// 然后通过我们之前写的 processMenuData 将 icon 字符串转为组件
	response.OkWithData(menus, c)
}

// GetMenuList 获取所有菜单
// @Tags Menu
// @Summary 分页获取基础menu列表 (虽然名字叫分页，实际上通常返回全量树)
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]system.SysMenu}
// @Router /menu/getMenuList [post]
func (a *MenuApi) GetMenuList(c *gin.Context) {
	// GVA 习惯用 POST 做查询，RESTful 建议 GET，这里兼容你的前端 API 定义
	menus, err := a.menuService.GetMenuList(c.Request.Context())
	if err != nil {
		a.svcCtx.Logger.Error("get_menu_list_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(menus, c)
}

// AddBaseMenu 新增菜单
// @Tags Menu
// @Summary 新增菜单
// @Security ApiKeyAuth
// @Param data body request.AddMenuReq true "参数"
// @Router /menu/addBaseMenu [post]
func (a *MenuApi) AddBaseMenu(c *gin.Context) {
	var req systemReq.AddMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := a.menuService.AddBaseMenu(c.Request.Context(), req); err != nil {
		a.svcCtx.Logger.Error("add_menu_error", zap.Error(err))
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// DeleteBaseMenu 删除菜单
// @Tags Menu
// @Summary 删除菜单
// @Security ApiKeyAuth
// @Param data body request.GetByIdReq true "ID"
// @Router /menu/deleteBaseMenu [post]
func (a *MenuApi) DeleteBaseMenu(c *gin.Context) {
	var req request.GetByIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := a.menuService.DeleteBaseMenu(c.Request.Context(), req.Uint()); err != nil {
		a.svcCtx.Logger.Error("delete_menu_error", zap.Error(err))
		// 直接把 Service 层的错误（如“存在子菜单”）返回给前端
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateBaseMenu 更新菜单
// @Tags Menu
// @Summary 更新菜单
// @Security ApiKeyAuth
// @Param data body request.UpdateMenuReq true "参数"
// @Router /menu/updateBaseMenu [post]
func (a *MenuApi) UpdateBaseMenu(c *gin.Context) {
	var req systemReq.UpdateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := a.menuService.UpdateBaseMenu(c.Request.Context(), req); err != nil {
		a.svcCtx.Logger.Error("update_menu_error", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetMenuAuthority 获取指定角色的菜单
// @Tags Menu
// @Summary 获取指定角色的菜单
// @Security ApiKeyAuth
// @Param data body request.GetAuthorityIdReq true "参数"
// @Router /menu/getMenuAuthority [post]
func (a *MenuApi) GetMenuAuthority(c *gin.Context) {
	var req systemReq.GetAuthorityIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	menus, err := a.menuService.GetMenuAuthority(c.Request.Context(), req.AuthorityId)
	if err != nil {
		a.svcCtx.Logger.Error("get_menu_authority_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}

	// 直接返回菜单数组
	response.OkWithData(menus, c)
}
