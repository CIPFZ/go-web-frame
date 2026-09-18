package magnet

import (
	"net/http"
	"os"
	"strings"

	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

type API struct {
	service *Service
}

type PreviewRequest struct {
	Magnet string `json:"magnet" binding:"required"`
}

func NewAPI(service *Service) *API {
	return &API{service: service}
}

func (a *API) Preview(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32768)
	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请输入有效的磁力链接", c)
		return
	}
	preview, err := a.service.Preview(c.Request.Context(), req.Magnet)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if preview.Cover != nil {
		preview.Cover.URL = strings.TrimSuffix(c.Request.URL.Path, "/preview") + "/cover/" + preview.InfoHash
	}
	response.OkWithData(preview, c)
}

func (a *API) Cover(c *gin.Context) {
	path, mime, err := a.service.ServeCover(c.Param("hash"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, max-age=3600")
	c.DataFromReader(http.StatusOK, -1, mime, file, nil)
}
