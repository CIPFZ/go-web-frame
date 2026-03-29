package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/dto"
	pluginModel "github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/repository"
	systemModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	systemRepo "github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"gorm.io/gorm"
)

const (
	pluginAuthorityRequester uint = 10010
	pluginAuthorityReviewer  uint = 10013
	pluginAuthorityPublisher uint = 10014
)

type IPluginService interface {
	GetPluginList(ctx context.Context, req dto.SearchPluginReq, userID uint) (common.PageResult, error)
	GetPluginOverview(ctx context.Context) (*dto.PluginOverview, error)
	GetProjectDetail(ctx context.Context, id uint) (*dto.ProjectDetail, error)
	GetPublishedPluginList(ctx context.Context, req dto.SearchPublishedPluginReq) (common.PageResult, error)
	GetPublishedPluginDetail(ctx context.Context, pluginID uint) (*dto.PublishedPluginDetail, error)
	CreatePlugin(ctx context.Context, req dto.CreatePluginReq, creatorID uint) error
	UpdatePlugin(ctx context.Context, req dto.UpdatePluginReq, actorID uint) error
	GetReleaseList(ctx context.Context, req dto.SearchReleaseReq) (common.PageResult, error)
	GetReleaseDetail(ctx context.Context, id uint) (*dto.ReleaseDetail, error)
	CreateRelease(ctx context.Context, req dto.CreateReleaseReq, creatorID uint) error
	UpdateRelease(ctx context.Context, req dto.UpdateReleaseReq) error
	TransitRelease(ctx context.Context, req dto.ReleaseActionReq, reviewerID uint) error
}

type PluginService struct {
	svcCtx *svc.ServiceContext
	repo   repository.IPluginRepository
	users  systemRepo.IUserRepository
}

func NewPluginService(svcCtx *svc.ServiceContext, repo repository.IPluginRepository, users systemRepo.IUserRepository) IPluginService {
	return &PluginService{svcCtx: svcCtx, repo: repo, users: users}
}

func (s *PluginService) GetPluginList(ctx context.Context, req dto.SearchPluginReq, userID uint) (common.PageResult, error) {
	if userID > 0 {
		user, err := s.users.FindById(ctx, userID)
		if err != nil {
			return common.PageResult{}, errors.New("current user not found")
		}
		if s.shouldScopePluginListToRequester(user) {
			req.CreatedBy = userID
		}
	}
	list, total, err := s.repo.ListPlugins(ctx, req)
	if err != nil {
		return common.PageResult{}, err
	}
	return common.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *PluginService) GetPluginOverview(ctx context.Context) (*dto.PluginOverview, error) {
	return s.repo.GetPluginOverview(ctx)
}

func (s *PluginService) GetProjectDetail(ctx context.Context, id uint) (*dto.ProjectDetail, error) {
	return s.repo.GetProjectDetail(ctx, id)
}

func (s *PluginService) GetPublishedPluginList(ctx context.Context, req dto.SearchPublishedPluginReq) (common.PageResult, error) {
	list, total, err := s.repo.ListPublishedPlugins(ctx, req)
	if err != nil {
		return common.PageResult{}, err
	}
	return common.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *PluginService) GetPublishedPluginDetail(ctx context.Context, pluginID uint) (*dto.PublishedPluginDetail, error) {
	return s.repo.GetPublishedPluginDetail(ctx, pluginID)
}

func (s *PluginService) CreatePlugin(ctx context.Context, req dto.CreatePluginReq, creatorID uint) error {
	if err := s.validatePluginReq(ctx, req, 0); err != nil {
		return err
	}
	plugin := &pluginModel.Plugin{
		Code:          strings.TrimSpace(req.Code),
		RepositoryURL: strings.TrimSpace(req.RepositoryURL),
		NameZh:        strings.TrimSpace(req.NameZh),
		NameEn:        strings.TrimSpace(req.NameEn),
		DescriptionZh: strings.TrimSpace(req.DescriptionZh),
		DescriptionEn: strings.TrimSpace(req.DescriptionEn),
		CapabilityZh:  strings.TrimSpace(req.CapabilityZh),
		CapabilityEn:  strings.TrimSpace(req.CapabilityEn),
		Owner:         strings.TrimSpace(req.Owner),
		CreatedBy:     creatorID,
		CurrentStatus: pluginModel.PluginStatusPlanning,
	}
	return s.repo.CreatePlugin(ctx, plugin)
}

func (s *PluginService) UpdatePlugin(ctx context.Context, req dto.UpdatePluginReq, actorID uint) error {
	plugin, err := s.repo.FindPluginByID(ctx, req.ID)
	if err != nil {
		return errors.New("plugin not found")
	}
	if plugin.CreatedBy == 0 {
		return errors.New("plugin owner is not configured")
	}
	if actorID == 0 || plugin.CreatedBy != actorID {
		return errors.New("you are not allowed to edit this project")
	}
	if err := s.validatePluginReq(ctx, req.CreatePluginReq, req.ID); err != nil {
		return err
	}
	plugin.Code = strings.TrimSpace(req.Code)
	plugin.RepositoryURL = strings.TrimSpace(req.RepositoryURL)
	plugin.NameZh = strings.TrimSpace(req.NameZh)
	plugin.NameEn = strings.TrimSpace(req.NameEn)
	plugin.DescriptionZh = strings.TrimSpace(req.DescriptionZh)
	plugin.DescriptionEn = strings.TrimSpace(req.DescriptionEn)
	plugin.CapabilityZh = strings.TrimSpace(req.CapabilityZh)
	plugin.CapabilityEn = strings.TrimSpace(req.CapabilityEn)
	plugin.Owner = strings.TrimSpace(req.Owner)
	if req.CurrentStatus != "" {
		plugin.CurrentStatus = req.CurrentStatus
	}
	return s.repo.UpdatePlugin(ctx, plugin)
}

func (s *PluginService) GetReleaseList(ctx context.Context, req dto.SearchReleaseReq) (common.PageResult, error) {
	list, total, err := s.repo.ListReleases(ctx, req)
	if err != nil {
		return common.PageResult{}, err
	}
	return common.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *PluginService) GetReleaseDetail(ctx context.Context, id uint) (*dto.ReleaseDetail, error) {
	return s.repo.GetReleaseDetail(ctx, id)
}

func (s *PluginService) CreateRelease(ctx context.Context, req dto.CreateReleaseReq, creatorID uint) error {
	if _, err := s.repo.FindPluginByID(ctx, req.PluginID); err != nil {
		return errors.New("plugin not found")
	}
	if err := s.validateReleaseReq(ctx, req, 0); err != nil {
		return err
	}
	checklist, err := repository.MarshalChecklist(req.Checklist)
	if err != nil {
		return err
	}
	release := &pluginModel.PluginRelease{
		PluginID:             req.PluginID,
		RequestType:          req.RequestType,
		Status:               initialReleaseStatus(req.RequestType),
		SourceReleaseID:      req.SourceReleaseID,
		TargetReleaseID:      req.TargetReleaseID,
		Version:              strings.TrimSpace(req.Version),
		VersionConstraint:    strings.TrimSpace(req.VersionConstraint),
		Publisher:            strings.TrimSpace(req.Publisher),
		ReviewerID:           req.ReviewerID,
		PublisherID:          req.PublisherID,
		Checklist:            checklist,
		PerformanceSummaryZh: strings.TrimSpace(req.PerformanceSummaryZh),
		PerformanceSummaryEn: strings.TrimSpace(req.PerformanceSummaryEn),
		TestReportURL:        strings.TrimSpace(req.TestReportURL),
		PackageX86URL:        strings.TrimSpace(req.PackageX86URL),
		PackageArmURL:        strings.TrimSpace(req.PackageArmURL),
		ChangelogZh:          strings.TrimSpace(req.ChangelogZh),
		ChangelogEn:          strings.TrimSpace(req.ChangelogEn),
		OfflineReasonZh:      strings.TrimSpace(req.OfflineReasonZh),
		OfflineReasonEn:      strings.TrimSpace(req.OfflineReasonEn),
		CreatedBy:            creatorID,
	}
	return s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(release).Error; err != nil {
			return err
		}
		return s.appendReleaseEvent(tx, release, "", release.Status, "create", creatorID, "ticket created")
	})
}

func (s *PluginService) UpdateRelease(ctx context.Context, req dto.UpdateReleaseReq) error {
	release, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return errors.New("release ticket not found")
	}
	if release.Status == pluginModel.PluginReleaseStatusReleased || release.Status == pluginModel.PluginReleaseStatusOfflined {
		return errors.New("released ticket is immutable")
	}
	if release.Status == pluginModel.PluginReleaseStatusPendingReview || release.Status == pluginModel.PluginReleaseStatusApproved {
		return errors.New("reviewing ticket cannot be edited")
	}
	if err := s.validateReleaseReq(ctx, req.CreateReleaseReq, req.ID); err != nil {
		return err
	}
	checklist, err := repository.MarshalChecklist(req.Checklist)
	if err != nil {
		return err
	}
	release.SourceReleaseID = req.SourceReleaseID
	release.TargetReleaseID = req.TargetReleaseID
	release.Version = strings.TrimSpace(req.Version)
	release.VersionConstraint = strings.TrimSpace(req.VersionConstraint)
	release.Publisher = strings.TrimSpace(req.Publisher)
	release.ReviewerID = req.ReviewerID
	release.PublisherID = req.PublisherID
	release.Checklist = checklist
	release.PerformanceSummaryZh = strings.TrimSpace(req.PerformanceSummaryZh)
	release.PerformanceSummaryEn = strings.TrimSpace(req.PerformanceSummaryEn)
	release.TestReportURL = strings.TrimSpace(req.TestReportURL)
	release.PackageX86URL = strings.TrimSpace(req.PackageX86URL)
	release.PackageArmURL = strings.TrimSpace(req.PackageArmURL)
	release.ChangelogZh = strings.TrimSpace(req.ChangelogZh)
	release.ChangelogEn = strings.TrimSpace(req.ChangelogEn)
	release.OfflineReasonZh = strings.TrimSpace(req.OfflineReasonZh)
	release.OfflineReasonEn = strings.TrimSpace(req.OfflineReasonEn)
	return s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Save(release).Error; err != nil {
			return err
		}
		return s.appendReleaseEvent(tx, release, release.Status, release.Status, "update", release.CreatedBy, "ticket updated")
	})
}

func (s *PluginService) TransitRelease(ctx context.Context, req dto.ReleaseActionReq, reviewerID uint) error {
	release, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return errors.New("release ticket not found")
	}
	now := time.Now()
	reviewComment := strings.TrimSpace(req.ReviewComment)

	return s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		fromStatus := release.Status
		switch req.Action {
		case dto.ReleaseActionSubmitReview:
			if err := s.ensureReleaseActor(ctx, reviewerID, &release.CreatedBy, pluginAuthorityRequester, nil, 0, release.ReviewerID, release.PublisherID); err != nil {
				return err
			}
			if release.Status != pluginModel.PluginReleaseStatusReleasePreparing && !(release.RequestType == pluginModel.PluginReleaseTypeOffline && release.Status == pluginModel.PluginReleaseStatusDraft) {
				return errors.New("current status cannot submit review")
			}
			if err := s.validateReviewSubmission(release); err != nil {
				return err
			}
			release.Status = pluginModel.PluginReleaseStatusPendingReview
			release.SubmittedAt = &now
		case dto.ReleaseActionApprove:
			if err := s.ensureReviewer(ctx, reviewerID, release); err != nil {
				return err
			}
			if release.Status != pluginModel.PluginReleaseStatusPendingReview {
				return errors.New("current status cannot approve")
			}
			release.Status = pluginModel.PluginReleaseStatusApproved
			release.ApprovedAt = &now
			release.ReviewedBy = &reviewerID
			release.ReviewComment = reviewComment
		case dto.ReleaseActionReject:
			if err := s.ensureReviewer(ctx, reviewerID, release); err != nil {
				return err
			}
			if release.Status != pluginModel.PluginReleaseStatusPendingReview {
				return errors.New("current status cannot reject")
			}
			if reviewComment == "" {
				return errors.New("review comment is required when rejecting")
			}
			release.Status = pluginModel.PluginReleaseStatusRejected
			release.ReviewedBy = &reviewerID
			release.ReviewComment = reviewComment
		case dto.ReleaseActionRevise:
			if err := s.ensureReleaseActor(ctx, reviewerID, &release.CreatedBy, pluginAuthorityRequester, nil, 0, release.ReviewerID, release.PublisherID); err != nil {
				return err
			}
			if release.Status != pluginModel.PluginReleaseStatusRejected {
				return errors.New("current status cannot revise")
			}
			if release.RequestType == pluginModel.PluginReleaseTypeOffline {
				release.Status = pluginModel.PluginReleaseStatusDraft
			} else {
				release.Status = pluginModel.PluginReleaseStatusReleasePreparing
			}
		case dto.ReleaseActionRelease:
			if err := s.ensurePublisher(ctx, reviewerID, release); err != nil {
				return err
			}
			if release.Status != pluginModel.PluginReleaseStatusApproved {
				return errors.New("current status cannot be released")
			}
			if release.RequestType == pluginModel.PluginReleaseTypeOffline {
				return s.executeOffline(ctx, tx, release, now, reviewComment)
			}
			return s.executeRelease(ctx, tx, release, now, reviewComment)
		default:
			return errors.New("unsupported action")
		}
		if err := tx.Save(release).Error; err != nil {
			return err
		}
		if err := s.appendReleaseEvent(tx, release, fromStatus, release.Status, string(req.Action), reviewerID, reviewComment); err != nil {
			return err
		}
		return s.createSystemNoticeForTransition(tx, release, fromStatus, release.Status, reviewComment)
	})
}

func (s *PluginService) executeRelease(ctx context.Context, tx *gorm.DB, release *pluginModel.PluginRelease, now time.Time, reviewComment string) error {
	release.Status = pluginModel.PluginReleaseStatusReleased
	release.ReleasedAt = &now
	release.ReviewComment = reviewComment
	if err := tx.Save(release).Error; err != nil {
		return err
	}
	plugin, err := s.repo.FindPluginByID(ctx, release.PluginID)
	if err != nil {
		return err
	}
	plugin.CurrentStatus = pluginModel.PluginStatusActive
	plugin.LatestVersion = release.Version
	plugin.LastReleasedAt = &now
	return tx.Save(plugin).Error
}

func (s *PluginService) executeOffline(ctx context.Context, tx *gorm.DB, release *pluginModel.PluginRelease, now time.Time, reviewComment string) error {
	if release.TargetReleaseID == nil {
		return errors.New("offline request requires a target release")
	}
	target, err := s.repo.FindReleaseByID(ctx, *release.TargetReleaseID)
	if err != nil {
		return errors.New("target release not found")
	}
	if target.Status != pluginModel.PluginReleaseStatusReleased || target.IsOfflined {
		return errors.New("target release is not an active released version")
	}
	target.IsOfflined = true
	target.OfflinedAt = &now
	if err := tx.Save(target).Error; err != nil {
		return err
	}

	release.Status = pluginModel.PluginReleaseStatusOfflined
	release.OfflinedAt = &now
	release.ReviewComment = reviewComment
	if err := tx.Save(release).Error; err != nil {
		return err
	}

	activeCount, err := s.repo.CountActiveReleasedVersions(ctx, release.PluginID)
	if err != nil {
		return err
	}
	plugin, err := s.repo.FindPluginByID(ctx, release.PluginID)
	if err != nil {
		return err
	}
	if activeCount == 0 {
		plugin.CurrentStatus = pluginModel.PluginStatusOfflined
	}
	return tx.Save(plugin).Error
}

func (s *PluginService) appendReleaseEvent(tx *gorm.DB, release *pluginModel.PluginRelease, fromStatus, toStatus pluginModel.PluginReleaseStatus, action string, operatorID uint, comment string) error {
	snapshot, err := json.Marshal(release)
	if err != nil {
		return err
	}
	event := pluginModel.PluginReleaseEvent{
		ReleaseID:    release.ID,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		Action:       action,
		OperatorID:   operatorID,
		Comment:      comment,
		SnapshotJSON: snapshot,
	}
	return tx.Create(&event).Error
}

func (s *PluginService) createSystemNoticeForTransition(tx *gorm.DB, release *pluginModel.PluginRelease, fromStatus, toStatus pluginModel.PluginReleaseStatus, comment string) error {
	if release.CreatedBy == 0 || fromStatus == toStatus && comment == "" {
		return nil
	}
	title := "Plugin workflow updated"
	content := "Plugin ticket #" + strconv.FormatUint(uint64(release.ID), 10) + " status changed to " + string(toStatus) + "."
	switch toStatus {
	case pluginModel.PluginReleaseStatusPendingReview:
		title = "Plugin ticket submitted for review"
		content = "Ticket #" + strconv.FormatUint(uint64(release.ID), 10) + " has entered review."
	case pluginModel.PluginReleaseStatusApproved:
		title = "Plugin ticket approved"
		content = "Ticket #" + strconv.FormatUint(uint64(release.ID), 10) + " was approved."
	case pluginModel.PluginReleaseStatusRejected:
		title = "Plugin ticket rejected"
		content = "Ticket #" + strconv.FormatUint(uint64(release.ID), 10) + " was rejected."
	case pluginModel.PluginReleaseStatusReleased:
		title = "Plugin version released"
		content = "Ticket #" + strconv.FormatUint(uint64(release.ID), 10) + " has been released."
	case pluginModel.PluginReleaseStatusOfflined:
		title = "Plugin version offlined"
		content = "Ticket #" + strconv.FormatUint(uint64(release.ID), 10) + " has been offlined."
	}
	if comment != "" {
		content += " Comment: " + comment
	}
	notice := systemModel.SysNotice{
		Title:       title,
		Content:     content,
		Level:       systemModel.NoticeLevelInfo,
		TargetType:  systemModel.NoticeTargetUsers,
		NeedConfirm: false,
		IsPopup:     false,
		CreatedBy:   release.CreatedBy,
	}
	if err := tx.Create(&notice).Error; err != nil {
		return err
	}
	receiver := systemModel.SysNoticeReceiver{
		NoticeID: notice.ID,
		UserID:   release.CreatedBy,
	}
	return tx.Create(&receiver).Error
}

func (s *PluginService) validatePluginReq(ctx context.Context, req dto.CreatePluginReq, excludeID uint) error {
	if strings.TrimSpace(req.Code) == "" ||
		strings.TrimSpace(req.RepositoryURL) == "" ||
		strings.TrimSpace(req.NameZh) == "" ||
		strings.TrimSpace(req.NameEn) == "" ||
		strings.TrimSpace(req.DescriptionZh) == "" ||
		strings.TrimSpace(req.DescriptionEn) == "" ||
		strings.TrimSpace(req.CapabilityZh) == "" ||
		strings.TrimSpace(req.CapabilityEn) == "" ||
		strings.TrimSpace(req.Owner) == "" {
		return errors.New("all bilingual plugin fields are required")
	}
	if _, err := s.repo.FindPluginByCode(ctx, strings.TrimSpace(req.Code), excludeID); err == nil {
		return errors.New("plugin code already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if _, err := s.repo.FindPluginByRepo(ctx, strings.TrimSpace(req.RepositoryURL), excludeID); err == nil {
		return errors.New("plugin repository already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func (s *PluginService) validateReleaseReq(ctx context.Context, req dto.CreateReleaseReq, excludeID uint) error {
	if _, err := s.repo.FindPluginByID(ctx, req.PluginID); err != nil {
		return errors.New("plugin not found")
	}
	switch req.RequestType {
	case pluginModel.PluginReleaseTypeInitial, pluginModel.PluginReleaseTypeMaintenance:
		if strings.TrimSpace(req.Version) == "" {
			return errors.New("version is required")
		}
		if _, err := s.repo.FindReleaseByVersion(ctx, req.PluginID, strings.TrimSpace(req.Version), excludeID); err == nil {
			return errors.New("version already exists for this plugin")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if req.RequestType == pluginModel.PluginReleaseTypeMaintenance && req.SourceReleaseID == nil {
			return errors.New("maintenance request requires a source release")
		}
	case pluginModel.PluginReleaseTypeOffline:
		if req.TargetReleaseID == nil {
			return errors.New("offline request requires a target release")
		}
		target, err := s.repo.FindReleaseByID(ctx, *req.TargetReleaseID)
		if err != nil {
			return errors.New("target release not found")
		}
		if target.PluginID != req.PluginID {
			return errors.New("target release does not belong to current plugin")
		}
		if target.Status != pluginModel.PluginReleaseStatusReleased || target.IsOfflined {
			return errors.New("target release is not an active released version")
		}
	default:
		return errors.New("invalid request type")
	}
	if req.ReviewerID != nil && req.PublisherID != nil && *req.ReviewerID == *req.PublisherID {
		return errors.New("reviewer and publisher must be different users")
	}
	if req.ReviewerID != nil && *req.ReviewerID == 0 || req.PublisherID != nil && *req.PublisherID == 0 {
		return errors.New("invalid assignee")
	}
	return nil
}

func initialReleaseStatus(requestType pluginModel.PluginReleaseType) pluginModel.PluginReleaseStatus {
	if requestType == pluginModel.PluginReleaseTypeOffline {
		return pluginModel.PluginReleaseStatusDraft
	}
	return pluginModel.PluginReleaseStatusReleasePreparing
}

func (s *PluginService) validateReviewSubmission(release *pluginModel.PluginRelease) error {
	if release.RequestType == pluginModel.PluginReleaseTypeOffline {
		if release.ReviewerID == nil || release.PublisherID == nil {
			return errors.New("offline request must assign reviewer and publisher")
		}
		if strings.TrimSpace(release.OfflineReasonZh) == "" || strings.TrimSpace(release.OfflineReasonEn) == "" {
			return errors.New("offline reason must be bilingual")
		}
		return nil
	}
	if release.ReviewerID == nil || release.PublisherID == nil {
		return errors.New("reviewer and publisher are required before review")
	}
	if strings.TrimSpace(release.Version) == "" ||
		strings.TrimSpace(release.Publisher) == "" ||
		strings.TrimSpace(release.TestReportURL) == "" ||
		strings.TrimSpace(release.PackageX86URL) == "" ||
		strings.TrimSpace(release.PackageArmURL) == "" ||
		strings.TrimSpace(release.ChangelogZh) == "" ||
		strings.TrimSpace(release.ChangelogEn) == "" {
		return errors.New("release package, report, publisher and bilingual changelog are required before review")
	}
	return nil
}

func (s *PluginService) ensureReleaseActor(ctx context.Context, userID uint, primaryAssignee *uint, primaryAuthority uint, secondaryAssignee *uint, secondaryAuthority uint, reviewerID *uint, publisherID *uint) error {
	if primaryAssignee != nil && *primaryAssignee == userID {
		return nil
	}
	if secondaryAssignee != nil && *secondaryAssignee == userID {
		return nil
	}
	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return errors.New("current user not found")
	}
	if hasAuthority(user, primaryAuthority) || (secondaryAuthority > 0 && hasAuthority(user, secondaryAuthority)) {
		if reviewerID != nil && *reviewerID == userID {
			return errors.New("reviewer cannot operate this step")
		}
		if publisherID != nil && *publisherID == userID && primaryAuthority != pluginAuthorityPublisher {
			return errors.New("publisher cannot operate this step")
		}
		return nil
	}
	return errors.New("you are not allowed to operate this step")
}

func (s *PluginService) ensureReviewer(ctx context.Context, userID uint, release *pluginModel.PluginRelease) error {
	if release.ReviewerID != nil && *release.ReviewerID == userID {
		if userID == release.CreatedBy {
			return errors.New("creator cannot self-review")
		}
		if release.PublisherID != nil && *release.PublisherID == userID {
			return errors.New("reviewer and publisher must be different")
		}
		return nil
	}
	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return errors.New("current user not found")
	}
	if !hasAuthority(user, pluginAuthorityReviewer) {
		return errors.New("only reviewer can approve or reject")
	}
	if userID == release.CreatedBy {
		return errors.New("creator cannot self-review")
	}
	if release.PublisherID != nil && *release.PublisherID == userID {
		return errors.New("reviewer and publisher must be different")
	}
	return nil
}

func (s *PluginService) ensurePublisher(ctx context.Context, userID uint, release *pluginModel.PluginRelease) error {
	if release.PublisherID != nil && *release.PublisherID == userID {
		if release.ReviewerID != nil && *release.ReviewerID == userID {
			return errors.New("reviewer and publisher must be different")
		}
		return nil
	}
	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return errors.New("current user not found")
	}
	if !hasAuthority(user, pluginAuthorityPublisher) {
		return errors.New("only publisher can execute release or offline")
	}
	if release.ReviewerID != nil && *release.ReviewerID == userID {
		return errors.New("reviewer and publisher must be different")
	}
	return nil
}

func hasAuthority(user *systemModel.SysUser, authorityID uint) bool {
	if user == nil || authorityID == 0 {
		return false
	}
	if user.AuthorityID == authorityID {
		return true
	}
	for _, item := range user.Authorities {
		if item.AuthorityId == authorityID {
			return true
		}
	}
	return false
}

func (s *PluginService) shouldScopePluginListToRequester(user *systemModel.SysUser) bool {
	if user == nil {
		return false
	}
	if !hasAuthority(user, pluginAuthorityRequester) {
		return false
	}
	authorityIDs := map[uint]struct{}{}
	if user.AuthorityID > 0 {
		authorityIDs[user.AuthorityID] = struct{}{}
	}
	for _, item := range user.Authorities {
		if item.AuthorityId > 0 {
			authorityIDs[item.AuthorityId] = struct{}{}
		}
	}
	if len(authorityIDs) != 1 {
		return false
	}
	_, ok := authorityIDs[pluginAuthorityRequester]
	return ok
}
