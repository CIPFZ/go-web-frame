package dto

import "github.com/CIPFZ/gowebframe/internal/modules/common"

type GenreReq struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sortOrder"`
}

type GenreSearchReq struct {
	common.PageInfo
	Name string `json:"name" form:"name"`
}
