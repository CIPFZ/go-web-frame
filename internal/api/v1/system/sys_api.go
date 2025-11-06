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

type SysApiApi struct {
	svcCtx        *svc.ServiceContext
	service       systemService.IApiService
	casbinService systemService.ICasbinService
}

func NewSysApiApi(svcCtx *svc.ServiceContext) *SysApiApi {
	return &SysApiApi{
		svcCtx:        svcCtx,
		service:       systemService.NewApiService(svcCtx),
		casbinService: systemService.NewCasbinService(svcCtx),
	}
}

// CreateApi 创建基础api
func (s *SysApiApi) CreateApi(c *gin.Context) {
	var api systemModel.SysApi
	log := logger.GetLogger(c)
	err := c.ShouldBindJSON(&api)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	err = s.service.CreateApi(api)
	if err != nil {
		log.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// SyncApi 同步API
func (s *SysApiApi) SyncApi(c *gin.Context) {
	log := logger.GetLogger(c)
	newApis, deleteApis, ignoreApis, err := s.service.SyncApi()
	if err != nil {
		log.Error("同步失败!", zap.Error(err))
		response.FailWithMessage("同步失败", c)
		return
	}
	response.OkWithData(gin.H{
		"newApis":    newApis,
		"deleteApis": deleteApis,
		"ignoreApis": ignoreApis,
	}, c)
}

// GetApiGroups 获取API分组
func (s *SysApiApi) GetApiGroups(c *gin.Context) {
	log := logger.GetLogger(c)
	groups, apiGroupMap, err := s.service.GetApiGroups()
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(gin.H{
		"groups":      groups,
		"apiGroupMap": apiGroupMap,
	}, c)
}

// IgnoreApi 忽略API
func (s *SysApiApi) IgnoreApi(c *gin.Context) {
	var ignoreApi systemModel.SysIgnoreApi
	err := c.ShouldBindJSON(&ignoreApi)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	err = s.service.IgnoreApi(ignoreApi)
	if err != nil {
		log.Error("忽略失败!", zap.Error(err))
		response.FailWithMessage("忽略失败", c)
		return
	}
	response.Ok(c)
}

// EnterSyncApi 确认同步API
func (s *SysApiApi) EnterSyncApi(c *gin.Context) {
	var syncApi systemRes.SysSyncApis
	err := c.ShouldBindJSON(&syncApi)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	err = s.service.EnterSyncApi(syncApi)
	if err != nil {
		log.Error("忽略失败!", zap.Error(err))
		response.FailWithMessage("忽略失败", c)
		return
	}
	response.Ok(c)
}

// DeleteApi 删除api
func (s *SysApiApi) DeleteApi(c *gin.Context) {
	var api systemModel.SysApi
	err := c.ShouldBindJSON(&api)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	err = s.service.DeleteApi(api)
	if err != nil {
		log.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetApiList 分页获取API列表
func (s *SysApiApi) GetApiList(c *gin.Context) {
	var pageInfo systemReq.SearchApiParams
	err := c.ShouldBindJSON(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	list, total, err := s.service.GetAPIInfoList(pageInfo.SysApi, pageInfo.PageInfo, pageInfo.OrderKey, pageInfo.Desc)
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

// GetApiById 根据id获取api
func (s *SysApiApi) GetApiById(c *gin.Context) {
	var idInfo request.GetById
	err := c.ShouldBindJSON(&idInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	api, err := s.service.GetApiById(idInfo.ID)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(systemRes.SysAPIResponse{Api: api}, "获取成功", c)
}

// UpdateApi 修改基础api
func (s *SysApiApi) UpdateApi(c *gin.Context) {
	var api systemModel.SysApi
	err := c.ShouldBindJSON(&api)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// TODO 参数校验
	log := logger.GetLogger(c)
	err = s.service.UpdateApi(api)
	if err != nil {
		log.Error("修改失败!", zap.Error(err))
		response.FailWithMessage("修改失败", c)
		return
	}
	response.OkWithMessage("修改成功", c)
}

// GetAllApis 获取所有的Api 不分页
func (s *SysApiApi) GetAllApis(c *gin.Context) {
	log := logger.GetLogger(c)
	authorityID := utils.GetUserAuthorityId(c, s.svcCtx)
	apis, err := s.service.GetAllApis(authorityID)
	if err != nil {
		log.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(systemRes.SysAPIListResponse{Apis: apis}, "获取成功", c)
}

// DeleteApisByIds 删除选中Api
func (s *SysApiApi) DeleteApisByIds(c *gin.Context) {
	var ids request.IdsReq
	err := c.ShouldBindJSON(&ids)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log := logger.GetLogger(c)
	err = s.service.DeleteApisByIds(ids)
	if err != nil {
		log.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FreshCasbin 刷新casbin缓存
func (s *SysApiApi) FreshCasbin(c *gin.Context) {
	log := logger.GetLogger(c)
	err := s.casbinService.FreshCasbin()
	if err != nil {
		log.Error("刷新失败!", zap.Error(err))
		response.FailWithMessage("刷新失败", c)
		return
	}
	response.OkWithMessage("刷新成功", c)
}
