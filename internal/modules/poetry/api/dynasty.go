package api

import (
	"strconv"

	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/dto"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (a *PoetryApi) CreateDynasty(c *gin.Context) {
	var req dto.DynastyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.service.CreateDynasty(c.Request.Context(), req); err != nil {
		log.Error("创建朝代失败", zap.Any("req", req), zap.Error(err))
		response.FailWithCode(errcode.CreateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) UpdateDynasty(c *gin.Context) {
	var req dto.DynastyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	log := logger.GetLogger(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if err := a.service.UpdateDynasty(c.Request.Context(), uint(id), req); err != nil {
		log.Error("更新朝代失败", zap.Int("id", id), zap.Any("req", req), zap.Error(err)) // ✨ 日志
		response.FailWithCode(errcode.UpdateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) DeleteDynasty(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	log := logger.GetLogger(c)
	if err := a.service.DeleteDynasty(c.Request.Context(), uint(id)); err != nil {
		// 将 Service 层返回的“该朝代下仍有诗人”错误透传给前端
		log.Warn("删除朝代失败", zap.Int("id", id), zap.Error(err))
		response.FailWithCode(errcode.DeleteFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) ListDynasty(c *gin.Context) {
	var req dto.DynastySearchReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	list, total, err := a.service.GetDynastyList(c.Request.Context(), req)
	if err != nil {
		response.FailWithCode(errcode.GetListFailed, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PoetryApi) AllDynasties(c *gin.Context) {
	list, err := a.service.GetAllDynasties(c.Request.Context())
	if err != nil {
		response.FailWithCode(errcode.GetDetailFailed, c)
		return
	}
	response.OkWithData(list, c)
}
