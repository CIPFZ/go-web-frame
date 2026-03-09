package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

type INoticeService interface {
	CreateNotice(ctx context.Context, req dto.CreateNoticeReq, creatorID uint) error
	GetNoticeList(ctx context.Context, req dto.SearchNoticeReq) ([]dto.NoticeListItem, int64, error)
	GetMyNotices(ctx context.Context, userID uint, page, pageSize int) ([]dto.MyNoticeItem, int64, error)
	MarkRead(ctx context.Context, noticeID uint, userID uint) error
}

type NoticeService struct {
	svcCtx     *svc.ServiceContext
	noticeRepo repository.INoticeRepository
}

func NewNoticeService(svcCtx *svc.ServiceContext, noticeRepo repository.INoticeRepository) INoticeService {
	return &NoticeService{svcCtx: svcCtx, noticeRepo: noticeRepo}
}

func (s *NoticeService) CreateNotice(ctx context.Context, req dto.CreateNoticeReq, creatorID uint) error {
	if req.Level == "" {
		req.Level = model.NoticeLevelInfo
	}
	if req.EndAt != nil && req.StartAt != nil && req.EndAt.Before(*req.StartAt) {
		return errors.New("endAt must be greater than startAt")
	}

	userIDs, err := s.resolveTargetUsers(ctx, req.TargetType, req.TargetIDs)
	if err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return errors.New("no valid receiver found")
	}

	notice := &model.SysNotice{
		Title:       req.Title,
		Content:     req.Content,
		Level:       req.Level,
		TargetType:  req.TargetType,
		IsPopup:     req.IsPopup,
		NeedConfirm: req.NeedConfirm,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		CreatedBy:   creatorID,
	}
	return s.noticeRepo.CreateWithReceivers(ctx, notice, userIDs)
}

func (s *NoticeService) GetNoticeList(ctx context.Context, req dto.SearchNoticeReq) ([]dto.NoticeListItem, int64, error) {
	return s.noticeRepo.GetNoticeList(ctx, req)
}

func (s *NoticeService) GetMyNotices(ctx context.Context, userID uint, page, pageSize int) ([]dto.MyNoticeItem, int64, error) {
	return s.noticeRepo.GetMyNotices(ctx, userID, page, pageSize)
}

func (s *NoticeService) MarkRead(ctx context.Context, noticeID uint, userID uint) error {
	return s.noticeRepo.MarkRead(ctx, noticeID, userID, time.Now())
}

func (s *NoticeService) resolveTargetUsers(ctx context.Context, targetType model.NoticeTargetType, targetIDs []uint) ([]uint, error) {
	var (
		ids []uint
		err error
	)

	switch targetType {
	case model.NoticeTargetAll:
		ids, err = s.noticeRepo.ListAllActiveUserIDs(ctx)
	case model.NoticeTargetRoles:
		ids, err = s.noticeRepo.ListUserIDsByAuthorityIDs(ctx, targetIDs)
	case model.NoticeTargetUsers:
		ids, err = s.noticeRepo.ListExistingUserIDs(ctx, targetIDs)
	default:
		return nil, errors.New("invalid targetType")
	}
	if err != nil {
		return nil, err
	}

	unique := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		unique[id] = struct{}{}
	}
	result := make([]uint, 0, len(unique))
	for id := range unique {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}
