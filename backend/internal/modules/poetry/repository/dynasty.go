package repository

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
)

func (r *PoetryRepo) CreateDynasty(ctx context.Context, d *model.MetaDynasty) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *PoetryRepo) UpdateDynasty(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.MetaDynasty{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PoetryRepo) DeleteDynasty(ctx context.Context, id uint) error {
	// 物理删除：因为 Model 中没有 DeletedAt，这里直接执行 DELETE 语句
	return r.db.WithContext(ctx).Delete(&model.MetaDynasty{}, id).Error
}

// HasAuthorsByDynasty 用于 Service 层在删除前进行安全校验
func (r *PoetryRepo) HasAuthorsByDynasty(ctx context.Context, dynastyId uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PoemAuthor{}).
		Where("dynasty_id = ?", dynastyId).
		Count(&count).Error
	return count > 0, err
}

func (r *PoetryRepo) GetDynastyList(ctx context.Context, page, size int, name string) (list []model.MetaDynasty, total int64, err error) {
	db := r.db.WithContext(ctx).Model(&model.MetaDynasty{})
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

func (r *PoetryRepo) GetAllDynasties(ctx context.Context) (list []model.MetaDynasty, err error) {
	err = r.db.WithContext(ctx).Order("sort_order desc").Find(&list).Error
	return
}
