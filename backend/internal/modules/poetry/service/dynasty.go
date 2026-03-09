package service

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
)

func (s *PoetryService) CreateDynasty(ctx context.Context, req dto.DynastyReq) error {
	return s.repo.CreateDynasty(ctx, &model.MetaDynasty{
		Name:      req.Name,
		SortOrder: req.SortOrder,
	})
}

func (s *PoetryService) UpdateDynasty(ctx context.Context, id uint, req dto.DynastyReq) error {
	// 使用 Map 进行更新，支持零值更新（如果需要）
	return s.repo.UpdateDynasty(ctx, id, map[string]interface{}{
		"name":       req.Name,
		"sort_order": req.SortOrder,
	})
}

func (s *PoetryService) DeleteDynasty(ctx context.Context, id uint) error {
	// 删除前检查引用
	// 即使数据库没有外键约束，业务逻辑上也必须阻止删除正在被使用的朝代
	has, err := s.repo.HasAuthorsByDynasty(ctx, id)
	if err != nil {
		return err
	}
	if has {
		return errcode.DynastyHasAuthors
	}

	return s.repo.DeleteDynasty(ctx, id)
}

func (s *PoetryService) GetDynastyList(ctx context.Context, req dto.DynastySearchReq) ([]model.MetaDynasty, int64, error) {
	return s.repo.GetDynastyList(ctx, req.Page, req.PageSize, req.Name)
}

func (s *PoetryService) GetAllDynasties(ctx context.Context) ([]model.MetaDynasty, error) {
	return s.repo.GetAllDynasties(ctx)
}
