package api

import (
	"fmt"
	
	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FileApi struct {
	svcCtx *svc.ServiceContext
}

func NewFileApi(svcCtx *svc.ServiceContext) *FileApi {
	return &FileApi{svcCtx: svcCtx}
}

// Upload 通用文件上传
// @Tags File
// @Summary 通用文件上传
// @Accept multipart/form-data
// @Param file formData file true "文件"
// @Success 200 {object} response.Response{data=string} "返回文件URL"
// @Router /sys/file/upload [post]
func (a *FileApi) Upload(c *gin.Context) {
	// 1. 获取文件
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择要上传的文件", c)
		return
	}

	log := logger.GetLogger(c)
	// 2. 获取当前操作人ID (仅用于记录日志，不绑定业务)
	uid := utils.GetUserUUID(c)

	// 3. 调用 Service 上传
	filename := fmt.Sprintf("%s/%s", uid, header.Filename)
	fileUrl, _, err := a.svcCtx.OSS.Upload(c.Request.Context(), header, filename)
	if err != nil {
		log.Error("文件上传失败", zap.Error(err))
		response.FailWithError(errcode.FileUploadFailed, c)
		return
	}

	// 4. 只返回 URL，不更新任何用户资料
	response.OkWithData(gin.H{"url": fileUrl}, c)
}
