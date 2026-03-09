package api

import (
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/service"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

type PoetryApi struct {
	svcCtx  *svc.ServiceContext
	service *service.PoetryService
}

func NewPoetryApi(svcCtx *svc.ServiceContext, service *service.PoetryService) *PoetryApi {
	return &PoetryApi{svcCtx: svcCtx, service: service}
}
