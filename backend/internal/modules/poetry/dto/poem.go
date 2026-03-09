package dto

import "github.com/CIPFZ/gowebframe/internal/modules/common"

type PoemReq struct {
	Title        string `json:"title" binding:"required"`
	AuthorID     uint   `json:"authorId" binding:"required"`
	GenreID      uint   `json:"genreId" binding:"required"`
	Content      string `json:"content" binding:"required"`
	Translation  string `json:"translation"`
	Annotation   string `json:"annotation"`
	Appreciation string `json:"appreciation"`
	AudioUrl     string `json:"audioUrl"`
	TagIds       []uint `json:"tagIds"`
}

type PoemSearchReq struct {
	common.PageInfo
	Keyword   string `json:"keyword" form:"keyword"`
	AuthorID  uint   `json:"authorId" form:"authorId"`
	GenreID   uint   `json:"genreId" form:"genreId"`
	DynastyID uint   `json:"dynastyId" form:"dynastyId"`
}
