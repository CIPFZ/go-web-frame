package media

import (
	"strconv"

	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service: service} }

type sourceRequest struct {
	Name                   string `json:"name" binding:"required"`
	APIURL                 string `json:"apiUrl" binding:"required"`
	APIType                string `json:"apiType"`
	Enabled                bool   `json:"enabled"`
	Priority               int    `json:"priority"`
	AllowCategories        string `json:"allowCategories"`
	DenyCategories         string `json:"denyCategories"`
	RequestIntervalSeconds int    `json:"requestIntervalSeconds"`
	TimeoutSeconds         int    `json:"timeoutSeconds"`
}

func sourceFromRequest(req sourceRequest) Source {
	return Source{Name: req.Name, APIURL: req.APIURL, APIType: req.APIType, Enabled: req.Enabled,
		Priority: req.Priority, AllowCategories: req.AllowCategories, DenyCategories: req.DenyCategories,
		RequestIntervalSeconds: req.RequestIntervalSeconds, TimeoutSeconds: req.TimeoutSeconds}
}

func (a *API) ListSources(c *gin.Context) {
	items, err := a.service.ListSources(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取资源站失败", c)
		return
	}
	response.OkWithData(items, c)
}

func (a *API) CreateSource(c *gin.Context) {
	var req sourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	item, err := a.service.CreateSource(c.Request.Context(), sourceFromRequest(req))
	if err != nil {
		response.FailWithMessage("创建资源站失败: "+err.Error(), c)
		return
	}
	response.OkWithData(item, c)
}

func (a *API) UpdateSource(c *gin.Context) {
	id, ok := parseUintID(c.Param("id"))
	if !ok {
		response.FailWithMessage("资源站 ID 无效", c)
		return
	}
	var req sourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	item := sourceFromRequest(req)
	item.ID = id
	updated, err := a.service.UpdateSource(c.Request.Context(), item)
	if err != nil {
		response.FailWithMessage("更新资源站失败: "+err.Error(), c)
		return
	}
	response.OkWithData(updated, c)
}

func (a *API) Health(c *gin.Context) {
	id, ok := parseUintID(c.Param("id"))
	if !ok {
		response.FailWithMessage("资源站 ID 无效", c)
		return
	}
	check, err := a.service.HealthCheck(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("健康检查失败: "+err.Error(), c)
		return
	}
	response.OkWithData(check, c)
}

func (a *API) Collect(c *gin.Context) {
	id, ok := parseUintID(c.Param("id"))
	if !ok {
		response.FailWithMessage("资源站 ID 无效", c)
		return
	}
	limit := 100
	if value, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil {
		limit = value
	}
	summary, err := a.service.Collect(c.Request.Context(), id, limit)
	if err != nil {
		response.FailWithMessage("采集失败: "+err.Error(), c)
		return
	}
	response.OkWithData(summary, c)
}

func (a *API) ListItems(c *gin.Context) {
	limit := 100
	if value, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil {
		limit = value
	}
	items, err := a.service.ListItems(c.Request.Context(), limit)
	if err != nil {
		response.FailWithMessage("获取影视条目失败", c)
		return
	}
	response.OkWithData(items, c)
}

func (a *API) GetItem(c *gin.Context) {
	id, ok := parseUintID(c.Param("id"))
	if !ok {
		response.FailWithMessage("影视 ID 无效", c)
		return
	}
	detail, err := a.service.GetItemDetail(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("影视条目不存在", c)
		return
	}
	response.OkWithData(detail, c)
}

func parseUintID(value string) (uint, bool) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	return uint(parsed), err == nil && parsed > 0
}
