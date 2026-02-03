package service

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
)

func (s *PoetryService) CreatePoem(ctx context.Context, req dto.PoemReq) error {
	work := &model.PoemWork{
		Title:        req.Title,
		AuthorID:     req.AuthorID,
		GenreID:      req.GenreID,
		Content:      req.Content,
		Translation:  req.Translation,
		Annotation:   req.Annotation,
		Appreciation: req.Appreciation,
		AudioUrl:     req.AudioUrl,
	}
	return s.repo.CreatePoem(ctx, work, req.TagIds)
}

func (s *PoetryService) UpdatePoem(ctx context.Context, id uint, req dto.PoemReq) error {
	// 注意：Repo 层的 UpdatePoem 接收的是 *model.PoemWork 结构体
	// 如果 req 中的某些字段为空字符串，GORM Updates 方法默认会忽略空值
	// 如果业务需求是“允许清空某个字段”，建议修改 Repo 层接受 map 或者使用 GORM 的 Select/Omit 功能
	// 暂时保持现状，通常文本字段部分更新是符合预期的
	work := &model.PoemWork{
		Title:        req.Title,
		AuthorID:     req.AuthorID,
		GenreID:      req.GenreID,
		Content:      req.Content,
		Translation:  req.Translation,
		Annotation:   req.Annotation,
		Appreciation: req.Appreciation,
		AudioUrl:     req.AudioUrl,
	}
	return s.repo.UpdatePoem(ctx, id, work, req.TagIds)
}

func (s *PoetryService) DeletePoem(ctx context.Context, id uint) error {
	// 作品删除是物理删除，Repo 层会自动清理关联表 (poem_tag_rel) 的数据
	return s.repo.DeletePoem(ctx, id)
}

func (s *PoetryService) GetPoemList(ctx context.Context, req dto.PoemSearchReq) ([]model.PoemWork, int64, error) {
	filter := map[string]interface{}{
		"author_id":  req.AuthorID,
		"genre_id":   req.GenreID,
		"dynasty_id": req.DynastyID,
		"keyword":    req.Keyword,
	}
	return s.repo.GetPoemList(ctx, req.Page, req.PageSize, filter)
}

func (s *PoetryService) GetPoemDetail(ctx context.Context, id uint) (*model.PoemWork, error) {
	return s.repo.GetPoemDetail(ctx, id)
}
