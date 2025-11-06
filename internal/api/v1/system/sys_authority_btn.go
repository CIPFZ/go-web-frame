package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemService "github.com/CIPFZ/gowebframe/internal/service/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthorityBtnApi struct {
	serviceCtx *svc.ServiceContext
	service    systemService.IAuthorityBtnService
}

func NewAuthorityBtnApi(svcCtx *svc.ServiceContext) *AuthorityBtnApi {
	return &AuthorityBtnApi{
		serviceCtx: svcCtx,
		service:    systemService.NewAuthorityBtnService(svcCtx),
	}
}

// GetAuthorityBtn 获取权限按钮
func (a *AuthorityBtnApi) GetAuthorityBtn(c *gin.Context) {
	var req systemReq.SysAuthorityBtnReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	res, err := a.service.GetAuthorityBtn(req)
	if err != nil {
		log.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithDetailed(res, "查询成功", c)
}

// SetAuthorityBtn 设置权限按钮
func (a *AuthorityBtnApi) SetAuthorityBtn(c *gin.Context) {
	var req systemReq.SysAuthorityBtnReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	err = a.service.SetAuthorityBtn(req)
	if err != nil {
		log.Error("分配失败!", zap.Error(err))
		response.FailWithMessage("分配失败", c)
		return
	}
	response.OkWithMessage("分配成功", c)
}

// CanRemoveAuthorityBtn 设置权限按钮
func (a *AuthorityBtnApi) CanRemoveAuthorityBtn(c *gin.Context) {
	id := c.Query("id")
	err := a.service.CanRemoveAuthorityBtn(id)
	log := logger.GetLogger(c)
	if err != nil {
		log.Error("删除失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
