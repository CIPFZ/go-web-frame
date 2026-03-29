package api

import (
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/service"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/gin-gonic/gin"
)

type PluginApi struct {
	svcCtx        *svc.ServiceContext
	pluginService service.IPluginService
}

func NewPluginApi(svcCtx *svc.ServiceContext, pluginService service.IPluginService) *PluginApi {
	return &PluginApi{svcCtx: svcCtx, pluginService: pluginService}
}

func (a *PluginApi) GetPluginList(c *gin.Context) {
	var req dto.SearchPluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	result, err := a.pluginService.GetPluginList(c.Request.Context(), req, utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage("get plugin list failed: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "ok", c)
}

func (a *PluginApi) GetPluginOverview(c *gin.Context) {
	result, err := a.pluginService.GetPluginOverview(c.Request.Context())
	if err != nil {
		response.FailWithMessage("get plugin overview failed: "+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *PluginApi) GetProjectDetail(c *gin.Context) {
	var req dto.GetProjectDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	result, err := a.pluginService.GetProjectDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.FailWithMessage("get project detail failed: "+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *PluginApi) GetPublishedPluginList(c *gin.Context) {
	var req dto.SearchPublishedPluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	result, err := a.pluginService.GetPublishedPluginList(c.Request.Context(), req)
	if err != nil {
		response.FailWithMessage("get published plugin list failed: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "ok", c)
}

func (a *PluginApi) GetPublishedPluginDetail(c *gin.Context) {
	var req dto.GetPublishedPluginDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	result, err := a.pluginService.GetPublishedPluginDetail(c.Request.Context(), req.PluginID)
	if err != nil {
		response.FailWithMessage("get published plugin detail failed: "+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *PluginApi) CreatePlugin(c *gin.Context) {
	var req dto.CreatePluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	if err := a.pluginService.CreatePlugin(c.Request.Context(), req, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("create plugin failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("created", c)
}

func (a *PluginApi) UpdatePlugin(c *gin.Context) {
	var req dto.UpdatePluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	if err := a.pluginService.UpdatePlugin(c.Request.Context(), req, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("update plugin failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("updated", c)
}

func (a *PluginApi) GetReleaseList(c *gin.Context) {
	var req dto.SearchReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	result, err := a.pluginService.GetReleaseList(c.Request.Context(), req)
	if err != nil {
		response.FailWithMessage("get release list failed: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "ok", c)
}

func (a *PluginApi) GetReleaseDetail(c *gin.Context) {
	var req dto.GetReleaseDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	result, err := a.pluginService.GetReleaseDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.FailWithMessage("get release detail failed: "+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *PluginApi) CreateRelease(c *gin.Context) {
	var req dto.CreateReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	if err := a.pluginService.CreateRelease(c.Request.Context(), req, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("create version workflow failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("created", c)
}

func (a *PluginApi) UpdateRelease(c *gin.Context) {
	var req dto.UpdateReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	if err := a.pluginService.UpdateRelease(c.Request.Context(), req); err != nil {
		response.FailWithMessage("update version workflow failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("updated", c)
}

func (a *PluginApi) TransitRelease(c *gin.Context) {
	var req dto.ReleaseActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	if err := a.pluginService.TransitRelease(c.Request.Context(), req, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("release transition failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("ok", c)
}
