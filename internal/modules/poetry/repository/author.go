package repository

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
)

func (r *PoetryRepo) CreateAuthor(ctx context.Context, a *model.PoemAuthor) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *PoetryRepo) UpdateAuthor(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.PoemAuthor{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PoetryRepo) DeleteAuthor(ctx context.Context, id uint) error {
	// 物理删除
	return r.db.WithContext(ctx).Delete(&model.PoemAuthor{}, id).Error
}

// HasWorksByAuthor 检查该诗人下是否有作品
func (r *PoetryRepo) HasWorksByAuthor(ctx context.Context, authorId uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PoemWork{}).
		Where("author_id = ?", authorId).
		Count(&count).Error
	return count > 0, err
}

func (r *PoetryRepo) GetAuthorList(ctx context.Context, page, size int, dynastyId uint, name string) (list []model.PoemAuthor, total int64, err error) {
	db := r.db.WithContext(ctx).Model(&model.PoemAuthor{}).Preload("Dynasty")
	if dynastyId > 0 {
		db = db.Where("dynasty_id = ?", dynastyId)
	}
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return
}

func (r *PoetryRepo) GetAuthorDetail(ctx context.Context, id uint) (a *model.PoemAuthor, err error) {
	err = r.db.WithContext(ctx).Preload("Dynasty").First(&a, id).Error
	return
}
