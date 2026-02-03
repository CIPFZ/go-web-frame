package api

import (
	"fmt"
	"strconv"

	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/dto"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func (a *PoetryApi) CreateAuthor(c *gin.Context) {
	var req dto.AuthorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.service.CreateAuthor(c.Request.Context(), req); err != nil {
		log.Error("创建诗人失败", zap.Any("req", req), zap.Error(err))
		response.FailWithCode(errcode.CreateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) UpdateAuthor(c *gin.Context) {
	var req dto.AuthorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.service.UpdateAuthor(c.Request.Context(), uint(id), req); err != nil {
		log.Error("更新诗人失败", zap.Int("id", id), zap.Any("req", req), zap.Error(err))
		response.FailWithCode(errcode.UpdateFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) DeleteAuthor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	log := logger.GetLogger(c)
	if err := a.service.DeleteAuthor(c.Request.Context(), uint(id)); err != nil {
		log.Warn("删除诗人失败", zap.Int("id", id), zap.Error(err))
		response.FailWithCode(errcode.DeleteFailed, c)
		return
	}
	response.Ok(c)
}

func (a *PoetryApi) ListAuthor(c *gin.Context) {
	var req dto.AuthorSearchReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	list, total, err := a.service.GetAuthorList(c.Request.Context(), req)
	if err != nil {
		response.FailWithCode(errcode.GetListFailed, c)
		return
	}
	response.OkWithPage(list, total, req.Page, req.PageSize, c)
}

func (a *PoetryApi) DetailAuthor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	data, err := a.service.GetAuthorDetail(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithCode(errcode.GetDetailFailed, c)
		return
	}
	response.OkWithData(data, c)
}

func (a *PoetryApi) UploadAuthorAvatar(c *gin.Context) {
	log := logger.GetLogger(c)
	idStr := c.PostForm("id")
	if idStr == "" {
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	authorId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error("诗人ID解析失败", zap.Error(err))
		response.FailWithCode(errcode.InvalidParams, c)
		return
	}
	// 1. 获取文件
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		response.FailWithCode(errcode.NoUploadFileFailed, c)
		return
	}

	// 3. 调用 Service 上传
	filename := fmt.Sprintf("poetry/author/avatar/%d", authorId)
	fileUrl, _, err := a.svcCtx.OSS.Upload(c.Request.Context(), header, filename)
	if err != nil {
		log.Error("文件上传失败", zap.Error(err))
		response.FailWithError(errcode.FileUploadFailed, c)
		return
	}
	// 更新author信息
	err = a.service.UpdateAuthorAvatar(c.Request.Context(), uint(authorId), fileUrl)
	if err != nil {
		log.Error("更新诗人头像失败", zap.Error(err))
		response.FailWithCode(errcode.UpdateFailed, c)
		return
	}
	// 4. 只返回 URL，不更新任何用户资料
	response.OkWithData(gin.H{"url": fileUrl}, c)
}
