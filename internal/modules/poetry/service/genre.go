package service

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
)

func (s *PoetryService) CreateGenre(ctx context.Context, req dto.GenreReq) error {
	return s.repo.CreateGenre(ctx, &model.MetaGenre{
		Name:      req.Name,
		SortOrder: req.SortOrder,
	})
}

func (s *PoetryService) UpdateGenre(ctx context.Context, id uint, req dto.GenreReq) error {
	return s.repo.UpdateGenre(ctx, id, map[string]interface{}{
		"name":       req.Name,
		"sort_order": req.SortOrder,
	})
}

func (s *PoetryService) DeleteGenre(ctx context.Context, id uint) error {
	// 删除前检查引用
	has, err := s.repo.HasWorksByGenre(ctx, id)
	if err != nil {
		return err
	}
	if has {
		return errcode.GenreHasWorks
	}

	return s.repo.DeleteGenre(ctx, id)
}

func (s *PoetryService) GetGenreList(ctx context.Context, req dto.GenreSearchReq) ([]model.MetaGenre, int64, error) {
	return s.repo.GetGenreList(ctx, req.Page, req.PageSize, req.Name)
}

func (s *PoetryService) GetAllGenres(ctx context.Context) ([]model.MetaGenre, error) {
	return s.repo.GetAllGenres(ctx)
}
