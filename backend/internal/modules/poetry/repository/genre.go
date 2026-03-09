package repository

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
)

func (r *PoetryRepo) CreateGenre(ctx context.Context, g *model.MetaGenre) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *PoetryRepo) UpdateGenre(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.MetaGenre{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PoetryRepo) DeleteGenre(ctx context.Context, id uint) error {
	// 物理删除
	return r.db.WithContext(ctx).Delete(&model.MetaGenre{}, id).Error
}

// HasWorksByGenre 检查该体裁下是否有作品
func (r *PoetryRepo) HasWorksByGenre(ctx context.Context, genreId uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PoemWork{}).
		Where("genre_id = ?", genreId).
		Count(&count).Error
	return count > 0, err
}

func (r *PoetryRepo) GetGenreList(ctx context.Context, page, size int, name string) (list []model.MetaGenre, total int64, err error) {
	db := r.db.WithContext(ctx).Model(&model.MetaGenre{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Order("sort_order desc, id asc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return
}

func (r *PoetryRepo) GetAllGenres(ctx context.Context) (list []model.MetaGenre, err error) {
	err = r.db.WithContext(ctx).Order("sort_order desc").Find(&list).Error
	return
}
