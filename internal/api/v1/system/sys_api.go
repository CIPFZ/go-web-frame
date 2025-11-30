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

type SysApiApi struct {
	svcCtx     *svc.ServiceContext
	apiService system.IApiService
}

func NewSysApiApi(svcCtx *svc.ServiceContext) *SysApiApi {
	return &SysApiApi{
		svcCtx:     svcCtx,
		apiService: system.NewApiService(svcCtx),
	}
}

// GetApiList 分页获取
func (a *SysApiApi) GetApiList(c *gin.Context) {
	var req request.SearchApiReq
	// 支持 JSON 或 Query 参数
	_ = c.ShouldBindJSON(&req)

	log := logger.GetLogger(c)
	list, total, err := a.apiService.GetApiList(c.Request.Context(), req)
	if err != nil {
		log.Error("get_api_list_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}

// CreateApi 新增
func (a *SysApiApi) CreateApi(c *gin.Context) {
	var req request.CreateApiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.apiService.CreateApi(c.Request.Context(), req); err != nil {
		log.Error("create_api_error", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateApi 更新
func (a *SysApiApi) UpdateApi(c *gin.Context) {
	var req request.UpdateApiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.apiService.UpdateApi(c.Request.Context(), req); err != nil {
		log.Error("update_api_error", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}

	// ✨ 关键：更新成功后，重新加载 Casbin 策略到内存
	// 因为我们在 Service 层修改了数据库中的规则，内存中还没更新
	if err := a.svcCtx.CasbinEnforcer.LoadPolicy(); err != nil {
		log.Error("reload_casbin_policy_error", zap.Error(err))
	}

	response.OkWithMessage("更新成功", c)
}

// DeleteApi 删除
func (a *SysApiApi) DeleteApi(c *gin.Context) {
	var req request.DeleteApiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.apiService.DeleteApi(c.Request.Context(), req); err != nil {
		log.Error("delete_api_error", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}

	// ✨ 关键：删除后也要刷新 Casbin 内存策略
	_ = a.svcCtx.CasbinEnforcer.LoadPolicy()

	response.OkWithMessage("删除成功", c)
}
