package system

import (
	"github.com/CIPFZ/gowebframe/internal/model/common/response"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OperationLogApi struct {
	svcCtx *svc.ServiceContext
}

func NewOperationLogApi(svcCtx *svc.ServiceContext) *OperationLogApi {
	return &OperationLogApi{svcCtx: svcCtx}
}

// GetOperationLogList 获取日志列表
// @Router /operationLog/getOperationLogList [post]
func (a *OperationLogApi) GetOperationLogList(c *gin.Context) {
	var req systemReq.SearchOperationLogReq
	_ = c.ShouldBindJSON(&req)

	// 调用接口方法 (ServiceContext 中存储的是接口)
	list, total, err := a.svcCtx.OperationLogService.GetOperationLogList(c.Request.Context(), req)
	if err != nil {
		a.svcCtx.Logger.Error("get_operation_log_list_error", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}

// DeleteOperationLogByIds 批量删除
// @Router /operationLog/deleteOperationLogByIds [post]
func (a *OperationLogApi) DeleteOperationLogByIds(c *gin.Context) {
	var req systemReq.DeleteOperationLogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := a.svcCtx.OperationLogService.DeleteOperationLogByIds(c.Request.Context(), req.IDs); err != nil {
		a.svcCtx.Logger.Error("delete_operation_log_error", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
