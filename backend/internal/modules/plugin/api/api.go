package api

import (
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/service"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type PluginApi struct {
	svcCtx  *svc.ServiceContext
	service service.IPluginService
}

func NewPluginApi(svcCtx *svc.ServiceContext, pluginService service.IPluginService) *PluginApi {
	return &PluginApi{svcCtx: svcCtx, service: pluginService}
}

func (a *PluginApi) GetPluginList(c *gin.Context) {
	var req dto.SearchPluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	list, total, err := a.service.GetPluginList(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PluginApi) GetProjectDetail(c *gin.Context) {
	var req dto.GetProjectDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.GetProjectDetail(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *PluginApi) CreatePlugin(c *gin.Context) {
	var req dto.CreatePluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.CreatePlugin(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *PluginApi) UpdatePlugin(c *gin.Context) {
	var req dto.UpdatePluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.UpdatePlugin(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *PluginApi) GetReleaseDetail(c *gin.Context) {
	var req dto.GetReleaseDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.GetReleaseDetail(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *PluginApi) CreateRelease(c *gin.Context) {
	var req dto.CreateReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.CreateRelease(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *PluginApi) UpdateRelease(c *gin.Context) {
	var req dto.UpdateReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.UpdateRelease(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *PluginApi) TransitionRelease(c *gin.Context) {
	var req dto.TransitionReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.TransitionRelease(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *PluginApi) ClaimWorkOrder(c *gin.Context) {
	var req dto.ClaimWorkOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.ClaimWorkOrder(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *PluginApi) ResetWorkOrder(c *gin.Context) {
	var req dto.ResetWorkOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.ResetWorkOrder(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}

func (a *PluginApi) GetWorkOrderPool(c *gin.Context) {
	var req dto.SearchWorkOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	list, total, err := a.service.GetWorkOrderPool(c.Request.Context(), c.GetUint("userId"), c.GetUint("authorityId"), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PluginApi) GetProductList(c *gin.Context) {
	var req dto.SearchProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	list, total, err := a.service.GetProductList(c.Request.Context(), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PluginApi) CreateProduct(c *gin.Context) {
	var req dto.CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.CreateProduct(c.Request.Context(), c.GetUint("authorityId"), req); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *PluginApi) UpdateProduct(c *gin.Context) {
	var req dto.UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	if err := a.service.UpdateProduct(c.Request.Context(), c.GetUint("authorityId"), req); err != nil {
		response.FailWithError(err, c)
		return
	}
	response.Ok(c)
}

func (a *PluginApi) GetDepartmentList(c *gin.Context) {
	var req dto.SearchDepartmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	list, total, err := a.service.GetDepartmentList(c.Request.Context(), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PluginApi) GetPublishedPluginList(c *gin.Context) {
	var req dto.GetPublishedPluginListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	list, total, err := a.service.GetPublishedPluginList(c.Request.Context(), req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PluginApi) GetPublishedPluginDetail(c *gin.Context) {
	var req dto.GetReleaseDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	item, err := a.service.GetPublishedPluginDetail(c.Request.Context(), req.ID)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	response.OkWithData(item, c)
}
