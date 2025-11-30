package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	"github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CasbinApi struct {
	svcCtx        *svc.ServiceContext
	casbinService system.ICasbinService
}

func NewCasbinApi(svcCtx *svc.ServiceContext) *CasbinApi {
	return &CasbinApi{
		svcCtx:        svcCtx,
		casbinService: system.NewCasbinService(svcCtx),
	}
}

// UpdateCasbin 更新角色API权限
// @Router /casbin/updateCasbin [post]
func (a *CasbinApi) UpdateCasbin(c *gin.Context) {
	var req request.UpdateCasbinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 简单校验
	if req.AuthorityId == "" {
		response.FailWithMessage("角色ID不能为空", c)
		return
	}
	log := logger.GetLogger(c)
	err := a.casbinService.UpdateCasbin(c.Request.Context(), req.AuthorityId, req.CasbinInfos)
	if err != nil {
		log.Error("update casbin error", zap.Error(err))
		response.FailWithMessage("更新权限失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetPolicyPathByAuthorityId 获取权限列表
// @Router /casbin/getPolicyPathByAuthorityId [post]
func (a *CasbinApi) GetPolicyPathByAuthorityId(c *gin.Context) {
	var req request.GetPolicyPathByAuthorityIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	list, err := a.casbinService.GetPolicyPathByAuthorityId(c.Request.Context(), req.AuthorityId)
	if err != nil {
		log.Error("get casbin policy error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(list, c)
}
