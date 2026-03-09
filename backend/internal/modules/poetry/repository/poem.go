package repository

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
	"gorm.io/gorm"
)

func (r *PoetryRepo) CreatePoem(ctx context.Context, w *model.PoemWork, tagIds []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(tagIds) > 0 {
			var tags []model.MetaTag
			if err := tx.Find(&tags, tagIds).Error; err != nil {
				return err
			}
			w.Tags = tags
		}
		return tx.Create(w).Error
	})
}

func (r *PoetryRepo) UpdatePoem(ctx context.Context, id uint, w *model.PoemWork, tagIds []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.PoemWork{}).Where("id = ?", id).Updates(w).Error; err != nil {
			return err
		}
		if tagIds != nil {
			var tags []model.MetaTag
			if len(tagIds) > 0 {
				if err := tx.Find(&tags, tagIds).Error; err != nil {
					return err
				}
			}
			var work model.PoemWork
			work.ID = id
			if err := tx.Model(&work).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PoetryRepo) DeletePoem(ctx context.Context, id uint) error {
	// ✨ 物理删除作品时，Select("Tags") 会自动清理 poem_tag_rel 表中的关联记录
	// 这样就不会在关联表中留下无效的 work_id
	return r.db.WithContext(ctx).Select("Tags").Delete(&model.PoemWork{}, id).Error
}

func (r *PoetryRepo) GetPoemList(ctx context.Context, page, size int, filter map[string]interface{}) (list []model.PoemWork, total int64, err error) {
	db := r.db.WithContext(ctx).Model(&model.PoemWork{}).
		Preload("Author").Preload("Genre").Preload("Author.Dynasty")

	if v, ok := filter["author_id"]; ok && v.(uint) > 0 {
		db = db.Where("author_id = ?", v)
	}
	if v, ok := filter["genre_id"]; ok && v.(uint) > 0 {
		db = db.Where("genre_id = ?", v)
	}
	if v, ok := filter["dynasty_id"]; ok && v.(uint) > 0 {
		db = db.Joins("JOIN poem_author ON poem_work.author_id = poem_author.id").
			Where("poem_author.dynasty_id = ?", v)
	}
	if v, ok := filter["keyword"]; ok && v.(string) != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+v.(string)+"%", "%"+v.(string)+"%")
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return
}

func (r *PoetryRepo) GetPoemDetail(ctx context.Context, id uint) (w *model.PoemWork, err error) {
	err = r.db.WithContext(ctx).
		Preload("Author").Preload("Genre").Preload("Author.Dynasty").Preload("Tags").
		First(&w, id).Error
	return
}
