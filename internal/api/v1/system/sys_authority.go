package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthorityApi struct {
	svcCtx      *svc.ServiceContext
	authService system.IAuthorityService
}

func NewAuthorityApi(svcCtx *svc.ServiceContext) *AuthorityApi {
	return &AuthorityApi{
		svcCtx:      svcCtx,
		authService: system.NewAuthorityService(svcCtx),
	}
}

// GetAuthorityList 获取角色列表
// @Router /authority/getAuthorityList [post]
func (a *AuthorityApi) GetAuthorityList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindJSON(&pageInfo) // 允许为空

	log := logger.GetLogger(c)
	list, total, err := a.authService.GetAuthorityList(c.Request.Context(), pageInfo)
	if err != nil {
		log.Error("get_authority_list_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     1,
		PageSize: int(total),
	}, "获取成功", c)
}

// CreateAuthority 创建角色
// @Router /authority/createAuthority [post]
func (a *AuthorityApi) CreateAuthority(c *gin.Context) {
	var req systemReq.CreateAuthorityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.authService.CreateAuthority(c.Request.Context(), req); err != nil {
		log.Error("create_authority_error", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateAuthority 更新角色
// @Router /authority/updateAuthority [post]
func (a *AuthorityApi) UpdateAuthority(c *gin.Context) {
	var req systemReq.UpdateAuthorityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.authService.UpdateAuthority(c.Request.Context(), req); err != nil {
		log.Error("update_authority_error", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteAuthority 删除角色
// @Router /authority/deleteAuthority [post]
func (a *AuthorityApi) DeleteAuthority(c *gin.Context) {
	var req request.GetByIdReq // 复用通用的ID请求结构 (id uint)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// 注意：request.GetByIdReq 的 ID 通常是自增ID，但这里我们需要的是 AuthorityId
	// 只要前端传参名为 "authorityId" 或 "id" 并能绑定到 uint 即可
	// 如果你的 DTO 字段叫 ID，前端传 authorityId 可能绑不上，建议专门定义一个 DeleteAuthorityReq { AuthorityId uint }
	log := logger.GetLogger(c)
	if err := a.authService.DeleteAuthority(c.Request.Context(), req.Uint()); err != nil {
		log.Error("delete_authority_error", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// SetAuthorityMenus 设置角色菜单权限
// @Router /authority/setDataAuthority [post]
func (a *AuthorityApi) SetAuthorityMenus(c *gin.Context) {
	var req systemReq.SetAuthorityMenusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.authService.SetAuthorityMenus(c.Request.Context(), req); err != nil {
		log.Error("set_authority_menu_error", zap.Error(err))
		response.FailWithMessage("设置失败", c)
		return
	}
	response.OkWithMessage("设置成功", c)
}
