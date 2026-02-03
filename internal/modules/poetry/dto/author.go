package dto

import "github.com/CIPFZ/gowebframe/internal/modules/common"

type AuthorReq struct {
	Name      string `json:"name" binding:"required"`
	DynastyID uint   `json:"dynastyId" binding:"required"`
	Intro     string `json:"intro"`
	LifeStory string `json:"lifeStory"`
	AvatarUrl string `json:"avatarUrl"`
}

type AuthorSearchReq struct {
	common.PageInfo
	Name      string `json:"name" form:"name"`
	DynastyID uint   `json:"dynastyId" form:"dynastyId"`
}
