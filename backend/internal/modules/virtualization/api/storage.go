package api

import (
	"github.com/CIPFZ/gowebframe/internal/modules/virtualization/service"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type StorageAPI struct {
	service *service.Service
}

func NewStorageAPI(s *service.Service) *StorageAPI {
	return &StorageAPI{service: s}
}

func (a *StorageAPI) ListPools(c *gin.Context) {
	pools, err := a.service.ListStoragePools()
	if err != nil {
		response.FailWithMessage("list storage pools failed: "+err.Error(), c)
		return
	}
	response.OkWithData(pools, c)
}

func (a *StorageAPI) ListVolumes(c *gin.Context) {
	volumes, err := a.service.ListStorageVolumes(c.Query("pool"))
	if err != nil {
		response.FailWithMessage("list storage volumes failed: "+err.Error(), c)
		return
	}
	response.OkWithData(volumes, c)
}

func (a *StorageAPI) CreateVolume(c *gin.Context) {
	var req service.CreateStorageVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	volume, err := a.service.CreateStorageVolume(req)
	if err != nil {
		response.FailWithMessage("create storage volume failed: "+err.Error(), c)
		return
	}
	response.OkWithData(volume, c)
}

func (a *StorageAPI) ResizeVolume(c *gin.Context) {
	var req service.ResizeStorageVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	volume, err := a.service.ResizeStorageVolume(c.Param("pool"), c.Param("name"), req)
	if err != nil {
		response.FailWithMessage("resize storage volume failed: "+err.Error(), c)
		return
	}
	response.OkWithData(volume, c)
}

func (a *StorageAPI) DeleteVolume(c *gin.Context) {
	if err := a.service.DeleteStorageVolume(c.Param("pool"), c.Param("name")); err != nil {
		response.FailWithMessage("delete storage volume failed: "+err.Error(), c)
		return
	}
	response.OkWithData(nil, c)
}
