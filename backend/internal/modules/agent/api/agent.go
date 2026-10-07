package api

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/CIPFZ/gowebframe/internal/modules/agent/service"
	virtualizationService "github.com/CIPFZ/gowebframe/internal/modules/virtualization/service"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type AgentAPI struct {
	gateway *service.GatewayClient
	vms     *virtualizationService.Service
}

func NewAgentAPI(gateway *service.GatewayClient, vms *virtualizationService.Service) *AgentAPI {
	return &AgentAPI{gateway: gateway, vms: vms}
}
func (a *AgentAPI) agentID(c *gin.Context) (string, error) {
	name := c.Param("name")
	if a.vms == nil {
		return name, nil
	}
	vm, err := a.vms.Get(name, false)
	if err != nil {
		return "", err
	}
	if vm.UUID == "" {
		return name, nil
	}
	return vm.UUID, nil
}

func (a *AgentAPI) Status(c *gin.Context) {
	agentID, err := a.agentID(c)
	var agent service.Agent
	if err == nil {
		agent, err = a.gateway.Agent(c.Request.Context(), agentID)
	}
	if err != nil {
		response.FailWithMessage("get agent status failed: "+err.Error(), c)
		return
	}
	response.OkWithData(agent, c)
}

type createTaskBody struct {
	CallerID       string         `json:"caller_id"`
	IdempotencyKey string         `json:"idempotency_key"`
	Action         map[string]any `json:"action"`
	TimeoutSeconds int64          `json:"timeout_seconds"`
	MaxAttempts    int            `json:"max_attempts"`
}

func (a *AgentAPI) CreateTask(c *gin.Context) {
	var req createTaskBody
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	if strings.TrimSpace(req.CallerID) == "" {
		req.CallerID = "cms"
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		req.IdempotencyKey = "cms-" + c.Param("name") + "-" + uuid.NewString()
	}
	if req.Action == nil {
		response.FailWithMessage("action is required", c)
		return
	}
	agentID, err := a.agentID(c)
	if err != nil {
		response.FailWithMessage("resolve VM agent failed: "+err.Error(), c)
		return
	}
	task, err := a.gateway.CreateTask(c.Request.Context(), agentID, req.CallerID, req.IdempotencyKey, req.Action, req.TimeoutSeconds, req.MaxAttempts)
	if err != nil {
		response.FailWithMessage("create agent task failed: "+err.Error(), c)
		return
	}
	response.OkWithData(task, c)
}
func (a *AgentAPI) Task(c *gin.Context) {
	task, err := a.gateway.Task(c.Request.Context(), c.Param("taskID"))
	if err != nil {
		response.FailWithMessage("get agent task failed: "+err.Error(), c)
		return
	}
	response.OkWithData(task, c)
}
func (a *AgentAPI) CancelTask(c *gin.Context) {
	task, err := a.gateway.Cancel(c.Request.Context(), c.Param("taskID"))
	if err != nil {
		response.FailWithMessage("cancel agent task failed: "+err.Error(), c)
		return
	}
	response.OkWithData(task, c)
}
func (a *AgentAPI) Health(c *gin.Context) {
	agentID, err := a.agentID(c)
	if err == nil {
		_, err = a.gateway.Agent(c.Request.Context(), agentID)
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	response.OkWithData(gin.H{"configured": true}, c)
}
