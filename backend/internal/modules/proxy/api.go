package proxy

import (
	"strconv"
	"strings"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service: service} }

type instanceRequest struct {
	Name        string `json:"name" binding:"required"`
	Engine      string `json:"engine" binding:"required"`
	Scope       string `json:"scope" binding:"required"`
	Unit        string `json:"unit" binding:"required"`
	BinaryPath  string `json:"binaryPath" binding:"required"`
	ConfigPath  string `json:"configPath" binding:"required"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type actionRequest struct {
	Action string `json:"action" binding:"required"`
}
type configRequest struct {
	Content string `json:"content" binding:"required"`
}

func (a *API) List(c *gin.Context) {
	items, err := a.service.List(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取代理实例失败", c)
		return
	}
	response.OkWithData(items, c)
}

func (a *API) Get(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	item, err := a.service.Get(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("实例不存在", c)
		return
	}
	response.OkWithData(instanceView(item, a.service.Status(c.Request.Context(), item)), c)
}

func (a *API) Create(c *gin.Context) {
	var req instanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	item, err := a.service.Create(c.Request.Context(), Instance{
		Name: req.Name, Engine: req.Engine, Scope: req.Scope, Unit: req.Unit,
		BinaryPath: req.BinaryPath, ConfigPath: req.ConfigPath, Enabled: req.Enabled, Description: req.Description,
	})
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(item, c)
}

func (a *API) Update(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	var req instanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	item, err := a.service.Update(c.Request.Context(), Instance{
		BaseModel: common.BaseModel{ID: id}, Name: req.Name, Engine: req.Engine, Scope: req.Scope,
		Unit: req.Unit, BinaryPath: req.BinaryPath, ConfigPath: req.ConfigPath,
		Enabled: req.Enabled, Description: req.Description,
	})
	if err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithData(item, c)
}

func (a *API) Action(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	var req actionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	status, err := a.service.Action(c.Request.Context(), id, strings.ToLower(req.Action))
	if err != nil {
		response.FailWithMessage("操作失败: "+err.Error(), c)
		return
	}
	response.OkWithData(status, c)
}

func (a *API) Validate(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	message, err := a.service.Validate(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("校验失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"message": message}, c)
}

func (a *API) ReadConfig(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	snapshot, err := a.service.ReadConfig(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("读取配置失败: "+err.Error(), c)
		return
	}
	response.OkWithData(snapshot, c)
}

func (a *API) SaveConfig(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	var req configRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	result, err := a.service.SaveConfig(c.Request.Context(), id, req.Content)
	if err != nil {
		response.FailWithMessage("保存配置失败: "+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *API) Rollback(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	result, err := a.service.Rollback(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("回滚失败: "+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *API) Metrics(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		response.FailWithMessage("实例 ID 无效", c)
		return
	}
	metrics, err := a.service.Metrics(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("获取指标失败", c)
		return
	}
	response.OkWithData(metrics, c)
}

func parseID(value string) (uint, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	return uint(parsed), err == nil && parsed > 0
}
