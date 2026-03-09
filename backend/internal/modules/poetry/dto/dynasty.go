package dto

import "github.com/CIPFZ/gowebframe/internal/modules/common"

type DynastyReq struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sortOrder"`
}

type DynastySearchReq struct {
	common.PageInfo
	Name string `json:"name" form:"name"`
}
