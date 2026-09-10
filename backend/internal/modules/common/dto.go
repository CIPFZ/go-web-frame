package common

import "gorm.io/gorm"

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page     int    `json:"page" form:"page"`         // 页码
	PageSize int    `json:"pageSize" form:"pageSize"` // 每页大小
	Keyword  string `json:"keyword" form:"keyword"`   // 关键字
}

// Normalize is shared by API responses and repository queries.
func (r *PageInfo) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 10
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	// Keep offset multiplication in range even for untrusted page numbers.
	maxPage := int(^uint(0)>>1) / r.PageSize
	if r.Page > maxPage {
		r.Page = maxPage
	}
}

func (r *PageInfo) Paginate() func(db *gorm.DB) *gorm.DB {
	r.Normalize()
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((r.Page - 1) * r.PageSize).Limit(r.PageSize)
	}
}

// GetByIdReq Find by id structure
type GetByIdReq struct {
	ID int `json:"id" form:"id"` // 主键ID
}

func (r *GetByIdReq) Uint() uint {
	return uint(r.ID)
}

type IdsReq struct {
	Ids []int `json:"ids" form:"ids"`
}

// GetAuthorityId Get role by id structure
type GetAuthorityId struct {
	AuthorityId uint `json:"authorityId" form:"authorityId"` // 角色ID
}

type Empty struct{}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}
