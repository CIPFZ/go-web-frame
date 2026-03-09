package repository

import (
	"context"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type INoticeRepository interface {
	CreateWithReceivers(ctx context.Context, notice *model.SysNotice, userIDs []uint) error
	GetNoticeList(ctx context.Context, req dto.SearchNoticeReq) ([]dto.NoticeListItem, int64, error)
	GetMyNotices(ctx context.Context, userID uint, page, pageSize int) ([]dto.MyNoticeItem, int64, error)
	MarkRead(ctx context.Context, noticeID uint, userID uint, readAt time.Time) error
	ListAllActiveUserIDs(ctx context.Context) ([]uint, error)
	ListUserIDsByAuthorityIDs(ctx context.Context, authorityIDs []uint) ([]uint, error)
	ListExistingUserIDs(ctx context.Context, userIDs []uint) ([]uint, error)
}

type NoticeRepository struct {
	db *gorm.DB
}

func NewNoticeRepository(db *gorm.DB) INoticeRepository {
	return &NoticeRepository{db: db}
}

func (r *NoticeRepository) CreateWithReceivers(ctx context.Context, notice *model.SysNotice, userIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return err
		}
		if len(userIDs) == 0 {
			return nil
		}
		receivers := make([]model.SysNoticeReceiver, 0, len(userIDs))
		for _, userID := range userIDs {
			receivers = append(receivers, model.SysNoticeReceiver{NoticeID: notice.ID, UserID: userID})
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&receivers).Error
	})
}

func (r *NoticeRepository) GetNoticeList(ctx context.Context, req dto.SearchNoticeReq) ([]dto.NoticeListItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	base := r.db.WithContext(ctx).Model(&model.SysNotice{})
	if req.Title != "" {
		base = base.Where("title LIKE ?", "%"+req.Title+"%")
	}
	if req.Level != "" {
		base = base.Where("level = ?", req.Level)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []dto.NoticeListItem
	err := base.Select(`
		sys_notices.id,
		sys_notices.created_at,
		sys_notices.title,
		sys_notices.content,
		sys_notices.level,
		sys_notices.target_type,
		sys_notices.is_popup,
		sys_notices.need_confirm,
		sys_notices.start_at,
		sys_notices.end_at,
		sys_notices.created_by,
		COUNT(DISTINCT nr.id) AS receiver_count,
		COUNT(DISTINCT CASE WHEN nr.read_at IS NOT NULL THEN nr.id END) AS read_count
	`).
		Joins("LEFT JOIN sys_notice_receivers nr ON nr.notice_id = sys_notices.id").
		Group("sys_notices.id").
		Order("sys_notices.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&list).Error
	return list, total, err
}

func (r *NoticeRepository) GetMyNotices(ctx context.Context, userID uint, page, pageSize int) ([]dto.MyNoticeItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	now := time.Now()

	base := r.db.WithContext(ctx).Table("sys_notice_receivers AS nr").
		Joins("JOIN sys_notices n ON n.id = nr.notice_id").
		Where("nr.user_id = ?", userID).
		Where("n.deleted_at IS NULL").
		Where("(n.start_at IS NULL OR n.start_at <= ?)", now).
		Where("(n.end_at IS NULL OR n.end_at >= ?)", now)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []dto.MyNoticeItem
	err := base.Select(`
		n.id,
		n.created_at,
		n.title,
		n.content,
		n.level,
		n.is_popup,
		n.need_confirm,
		n.start_at,
		n.end_at,
		nr.read_at
	`).
		Order("CASE WHEN nr.read_at IS NULL THEN 0 ELSE 1 END ASC, n.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}

func (r *NoticeRepository) MarkRead(ctx context.Context, noticeID uint, userID uint, readAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.SysNoticeReceiver{}).
		Where("notice_id = ? AND user_id = ?", noticeID, userID).
		Where("read_at IS NULL").
		Update("read_at", readAt).Error
}

func (r *NoticeRepository) ListAllActiveUserIDs(ctx context.Context) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&model.SysUser{}).Where("status = ?", model.UserActive).Pluck("id", &ids).Error
	return ids, err
}

func (r *NoticeRepository) ListUserIDsByAuthorityIDs(ctx context.Context, authorityIDs []uint) ([]uint, error) {
	if len(authorityIDs) == 0 {
		return []uint{}, nil
	}
	var ids []uint
	err := r.db.WithContext(ctx).
		Table("sys_user_authorities AS ua").
		Distinct("ua.user_id").
		Joins("JOIN sys_users u ON u.id = ua.user_id").
		Where("ua.authority_id IN ?", authorityIDs).
		Where("u.status = ?", model.UserActive).
		Pluck("ua.user_id", &ids).Error
	return ids, err
}

func (r *NoticeRepository) ListExistingUserIDs(ctx context.Context, userIDs []uint) ([]uint, error) {
	if len(userIDs) == 0 {
		return []uint{}, nil
	}
	var ids []uint
	err := r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("id IN ?", userIDs).
		Where("status = ?", model.UserActive).
		Pluck("id", &ids).Error
	return ids, err
}
