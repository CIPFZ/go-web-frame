package system

import (
	"github.com/CIPFZ/gowebframe/internal/api/v1/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

type OperationRecordRouter struct {
	serviceCtx *svc.ServiceContext
}

func (s *OperationRecordRouter) InitSysOperationRecordRouter(Router *gin.RouterGroup) {
	operationRecordRouter := Router.Group("sysOperationRecord")
	operationRecordApi := system.NewOperationRecordApi(s.serviceCtx)
	operationRecordRouter.DELETE("deleteSysOperationRecord", operationRecordApi.DeleteSysOperationRecord)           // 删除SysOperationRecord
	operationRecordRouter.DELETE("deleteSysOperationRecordByIds", operationRecordApi.DeleteSysOperationRecordByIds) // 批量删除SysOperationRecord
	operationRecordRouter.GET("findSysOperationRecord", operationRecordApi.FindSysOperationRecord)                  // 根据ID获取SysOperationRecord
	operationRecordRouter.POST("getSysOperationRecordList", operationRecordApi.GetSysOperationRecordList)           // 获取SysOperationRecord列表
}
