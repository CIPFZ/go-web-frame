package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CasbinApi struct {
	svcCtx *svc.ServiceContext
}

func NewCasbinApi(svcCtx *svc.ServiceContext) *CasbinApi {
	return &CasbinApi{svcCtx: svcCtx}
}

// UpdateCasbin 更新角色api权限
func (cas *CasbinApi) UpdateCasbin(c *gin.Context) {
	var cmr systemReq.CasbinInReceive
	err := c.ShouldBindJSON(&cmr)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	adminAuthorityID := utils.GetUserAuthorityId(c, cas.svcCtx)
	err = systemService.NewCasbinService(cas.svcCtx).UpdateCasbin(adminAuthorityID, cmr.AuthorityId, cmr.CasbinInfos)
	if err != nil {
		log.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetPolicyPathByAuthorityId 获取权限列表
func (cas *CasbinApi) GetPolicyPathByAuthorityId(c *gin.Context) {
	var casbin systemReq.CasbinInReceive
	err := c.ShouldBindJSON(&casbin)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	paths := systemService.NewCasbinService(cas.svcCtx).GetPolicyPathByAuthorityId(casbin.AuthorityId)
	response.OkWithDetailed(systemRes.PolicyPathResponse{Paths: paths}, "获取成功", c)
}
