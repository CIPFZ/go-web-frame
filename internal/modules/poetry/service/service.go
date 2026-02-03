package service

import (
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

type PoetryService struct {
	svcCtx *svc.ServiceContext
	repo   *repository.PoetryRepo
}

func NewPoetryService(svcCtx *svc.ServiceContext, repo *repository.PoetryRepo) *PoetryService {
	return &PoetryService{svcCtx: svcCtx, repo: repo}
}
