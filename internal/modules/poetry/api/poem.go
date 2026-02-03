package api

import (
	"strconv"

	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/dto"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func (a *PoetryApi) CreatePoem(c *gin.Context) {
	var req dto.PoemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.service.CreatePoem(c.Request.Context(), req); err != nil {
		log.Error("创建作品失败", zap.Any("req", req), zap.Error(err))
		response.FailWithCode(errcode.CreateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) UpdatePoem(c *gin.Context) {
	var req dto.PoemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	log := logger.GetLogger(c)
	if err := a.service.UpdatePoem(c.Request.Context(), uint(id), req); err != nil {
		log.Error("更新作品失败", zap.Int("id", id), zap.Any("req", req), zap.Error(err))
		response.FailWithCode(errcode.UpdateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) DeletePoem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	log := logger.GetLogger(c)
	if err := a.service.DeletePoem(c.Request.Context(), uint(id)); err != nil {
		log.Error("删除作品失败", zap.Int("id", id), zap.Error(err))
		response.FailWithCode(errcode.DeleteFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) ListPoem(c *gin.Context) {
	var req dto.PoemSearchReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	list, total, err := a.service.GetPoemList(c.Request.Context(), req)
	if err != nil {
		response.FailWithCode(errcode.GetListFailed, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PoetryApi) DetailPoem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := a.service.GetPoemDetail(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithCode(errcode.GetDetailFailed, c)
		return
	}
	response.OkWithData(data, c)
}
