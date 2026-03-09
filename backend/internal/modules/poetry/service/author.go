package service

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
)

func (s *PoetryService) CreateAuthor(ctx context.Context, req dto.AuthorReq) error {
	return s.repo.CreateAuthor(ctx, &model.PoemAuthor{
		Name:      req.Name,
		DynastyID: req.DynastyID,
		Intro:     req.Intro,
		LifeStory: req.LifeStory,
		AvatarUrl: req.AvatarUrl,
	})
}

func (s *PoetryService) UpdateAuthor(ctx context.Context, id uint, req dto.AuthorReq) error {
	// Map 更新模式：更加灵活，避免 Struct 更新时的零值忽略问题
	return s.repo.UpdateAuthor(ctx, id, map[string]interface{}{
		"name":       req.Name,
		"dynasty_id": req.DynastyID,
		"intro":      req.Intro,
		"life_story": req.LifeStory,
	})
}

func (s *PoetryService) UpdateAuthorAvatar(ctx context.Context, id uint, avatarUrl string) error {
	// Map 更新模式：更加灵活，避免 Struct 更新时的零值忽略问题
	return s.repo.UpdateAuthor(ctx, id, map[string]interface{}{
		"avatar_url": avatarUrl,
	})
}

func (s *PoetryService) DeleteAuthor(ctx context.Context, id uint) error {
	// 删除前检查引用
	has, err := s.repo.HasWorksByAuthor(ctx, id)
	if err != nil {
		return err
	}
	if has {
		return errcode.AuthorHasWorks
	}

	return s.repo.DeleteAuthor(ctx, id)
}

func (s *PoetryService) GetAuthorList(ctx context.Context, req dto.AuthorSearchReq) ([]model.PoemAuthor, int64, error) {
	return s.repo.GetAuthorList(ctx, req.Page, req.PageSize, req.DynastyID, req.Name)
}

func (s *PoetryService) GetAuthorDetail(ctx context.Context, id uint) (*model.PoemAuthor, error) {
	return s.repo.GetAuthorDetail(ctx, id)
}
