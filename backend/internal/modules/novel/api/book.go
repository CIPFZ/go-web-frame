package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/CIPFZ/gowebframe/internal/modules/novel/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookAPI struct{ db *gorm.DB }

func NewBookAPI(svcCtx *svc.ServiceContext) *BookAPI { return &BookAPI{db: svcCtx.DB} }

type listBooksRequest struct {
	common.PageInfo
	Author   string `json:"author" form:"author"`
	Category string `json:"category" form:"category"`
}

func (a *BookAPI) List(c *gin.Context) {
	var req listBooksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithValidation(err, c)
		return
	}
	req.Normalize()
	query := a.db.WithContext(c.Request.Context()).Model(&model.NovelBook{})
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR author LIKE ?", like, like)
	}
	if author := strings.TrimSpace(req.Author); author != "" {
		query = query.Where("author LIKE ?", "%"+author+"%")
	}
	if category := strings.TrimSpace(req.Category); category != "" {
		query = query.Where("category = ?", category)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.FailWithMessage("get novel books failed", c)
		return
	}
	list := make([]model.NovelBook, 0, req.PageSize)
	if err := query.Select("id, title, author, category, source_url, language, formats, created_at, updated_at").
		Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list).Error; err != nil {
		response.FailWithMessage("get novel books failed", c)
		return
	}
	response.OkWithDetailed(common.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

func (a *BookAPI) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("invalid novel book id", c)
		return
	}
	var book model.NovelBook
	if err := a.db.WithContext(c.Request.Context()).First(&book, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.FailWithMessage("novel book not found", c)
			return
		}
		response.FailWithMessage("get novel book failed", c)
		return
	}
	c.JSON(http.StatusOK, response.Response{Code: 0, Msg: "ok", Data: book})
}
