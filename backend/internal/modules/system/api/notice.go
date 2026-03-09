package api

import (
	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/service"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NoticeApi struct {
	svcCtx        *svc.ServiceContext
	noticeService service.INoticeService
}

func NewNoticeApi(svcCtx *svc.ServiceContext, noticeService service.INoticeService) *NoticeApi {
	return &NoticeApi{svcCtx: svcCtx, noticeService: noticeService}
}

func (a *NoticeApi) CreateNotice(c *gin.Context) {
	var req dto.CreateNoticeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	if err := a.noticeService.CreateNotice(c.Request.Context(), req, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("create notice failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("created", c)
}

func (a *NoticeApi) GetNoticeList(c *gin.Context) {
	var req dto.SearchNoticeReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	list, total, err := a.noticeService.GetNoticeList(c.Request.Context(), req)
	if err != nil {
		logger.GetLogger(c).Error("get_notice_list_failed", zap.Error(err))
		response.FailWithMessage("get notice list failed", c)
		return
	}
	response.OkWithDetailed(common.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "ok", c)
}

func (a *NoticeApi) GetMyNotices(c *gin.Context) {
	type query struct {
		Page     int `form:"page"`
		PageSize int `form:"pageSize"`
	}
	var q query
	_ = c.ShouldBindQuery(&q)
	list, total, err := a.noticeService.GetMyNotices(c.Request.Context(), utils.GetUserID(c), q.Page, q.PageSize)
	if err != nil {
		logger.GetLogger(c).Error("get_my_notices_failed", zap.Error(err))
		response.FailWithMessage("get my notices failed", c)
		return
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	response.OkWithDetailed(common.PageResult{
		List:     list,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, "ok", c)
}

func (a *NoticeApi) MarkRead(c *gin.Context) {
	var req dto.MarkNoticeReadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request: "+err.Error(), c)
		return
	}
	if err := a.noticeService.MarkRead(c.Request.Context(), req.NoticeID, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("mark read failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("ok", c)
}
