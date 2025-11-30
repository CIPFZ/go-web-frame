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
	operationRecordRouter := Router.Group("operationLog")
	operationRecordApi := system.NewOperationLogApi(s.serviceCtx)
	operationRecordRouter.POST("getOperationLogList", operationRecordApi.GetOperationLogList)         // 分页获取列表
	operationRecordRouter.POST("deleteOperationLogByIds", operationRecordApi.DeleteOperationLogByIds) // 批量删除
}
