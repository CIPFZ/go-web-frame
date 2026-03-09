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

func (a *PoetryApi) CreateGenre(c *gin.Context) {
	var req dto.GenreReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.service.CreateGenre(c.Request.Context(), req); err != nil {
		log.Error("创建体裁失败", zap.Any("req", req), zap.Error(err))
		response.FailWithCode(errcode.CreateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) UpdateGenre(c *gin.Context) {
	var req dto.GenreReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	log := logger.GetLogger(c)
	if err := a.service.UpdateGenre(c.Request.Context(), uint(id), req); err != nil {
		log.Error("更新体裁失败", zap.Int("id", id), zap.Any("req", req), zap.Error(err))
		// ✨ 返回具体错误信息 (如：名称重复)
		response.FailWithCode(errcode.UpdateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) DeleteGenre(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	log := logger.GetLogger(c)
	if err := a.service.DeleteGenre(c.Request.Context(), uint(id)); err != nil {
		// 关键：透传“该体裁下仍有作品”错误
		log.Warn("删除体裁失败", zap.Int("id", id), zap.Error(err))
		response.FailWithCode(errcode.DeleteFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) ListGenre(c *gin.Context) {
	var req dto.GenreSearchReq
	// 使用 ShouldBindQuery 自动解析 PageInfo
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	list, total, err := a.service.GetGenreList(c.Request.Context(), req)
	if err != nil {
		response.FailWithCode(errcode.GetListFailed, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PoetryApi) AllGenres(c *gin.Context) {
	list, err := a.service.GetAllGenres(c.Request.Context())
	if err != nil {
		response.FailWithCode(errcode.GetDetailFailed, c)
		return
	}
	response.OkWithData(list, c)
}
