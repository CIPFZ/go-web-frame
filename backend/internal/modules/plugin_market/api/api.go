package api

import (
	"strings"

	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/service"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type Api struct {
	service   *service.Service
	syncToken string
}

func New(service *service.Service, syncToken string) *Api {
	return &Api{service: service, syncToken: strings.TrimSpace(syncToken)}
}

func (a *Api) ListPlugins(c *gin.Context) {
	var req dto.ListPluginsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	list, total, err := a.service.ListPublishedPlugins(c.Request.Context(), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *Api) GetPluginDetail(c *gin.Context) {
	var req dto.GetPluginDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.GetPublishedPluginDetail(c.Request.Context(), req.PluginID)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *Api) UpsertPlugin(c *gin.Context) {
	if !a.authorizeSync(c) {
		return
	}
	var req dto.UpsertPluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.UpsertPlugin(c.Request.Context(), req); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *Api) UpsertVersion(c *gin.Context) {
	if !a.authorizeSync(c) {
		return
	}
	var req dto.UpsertVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.UpsertVersion(c.Request.Context(), req); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *Api) OfflineVersion(c *gin.Context) {
	if !a.authorizeSync(c) {
		return
	}
	var req dto.OfflineVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.OfflineVersion(c.Request.Context(), req.ReleaseID); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *Api) DeletePlugin(c *gin.Context) {
	if !a.authorizeSync(c) {
		return
	}
	var req dto.DeletePluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.DeletePlugin(c.Request.Context(), req.PluginID); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *Api) authorizeSync(c *gin.Context) bool {
	if a.syncToken == "" {
		response.FailWithMessage("sync token not configured", c)
		return false
	}
	if strings.TrimSpace(c.GetHeader("X-Market-Sync-Token")) != a.syncToken {
		response.FailWithMessage("unauthorized", c)
		return false
	}
	return true
}
