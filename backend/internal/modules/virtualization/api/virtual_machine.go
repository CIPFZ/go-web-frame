package api

import (
	"net/http"

	"github.com/CIPFZ/gowebframe/internal/modules/virtualization/service"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type VirtualMachineAPI struct {
	service *service.Service
}

func NewVirtualMachineAPI(s *service.Service) *VirtualMachineAPI {
	return &VirtualMachineAPI{service: s}
}

func (a *VirtualMachineAPI) List(c *gin.Context) {
	vms, err := a.service.List()
	if err != nil {
		response.FailWithMessage("list virtual machines failed: "+err.Error(), c)
		return
	}
	response.OkWithData(vms, c)
}

func (a *VirtualMachineAPI) Detail(c *gin.Context) {
	vm, err := a.service.Get(c.Param("name"), true)
	if err != nil {
		response.FailWithMessage("get virtual machine failed: "+err.Error(), c)
		return
	}
	response.OkWithData(vm, c)
}

func (a *VirtualMachineAPI) Create(c *gin.Context) {
	var req service.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	vm, err := a.service.Create(req)
	if err != nil {
		response.FailWithMessage("create virtual machine failed: "+err.Error(), c)
		return
	}
	response.OkWithData(vm, c)
}

func (a *VirtualMachineAPI) Update(c *gin.Context) {
	var req service.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	vm, err := a.service.Update(c.Param("name"), req)
	if err != nil {
		response.FailWithMessage("update virtual machine failed: "+err.Error(), c)
		return
	}
	response.OkWithData(vm, c)
}

func (a *VirtualMachineAPI) action(c *gin.Context, fn func(string) error) {
	if err := fn(c.Param("name")); err != nil {
		response.FailWithMessage("virtual machine action failed: "+err.Error(), c)
		return
	}
	c.JSON(http.StatusOK, response.Response{Code: 0, Msg: "ok"})
}

func (a *VirtualMachineAPI) Start(c *gin.Context) {
	a.action(c, a.service.Start)
}

func (a *VirtualMachineAPI) Shutdown(c *gin.Context) {
	a.action(c, a.service.Shutdown)
}

func (a *VirtualMachineAPI) Reboot(c *gin.Context) {
	a.action(c, a.service.Reboot)
}

func (a *VirtualMachineAPI) ForceStop(c *gin.Context) {
	a.action(c, a.service.ForceStop)
}

func (a *VirtualMachineAPI) Delete(c *gin.Context) {
	a.action(c, a.service.Delete)
}
