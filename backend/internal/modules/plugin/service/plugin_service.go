package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/plugin/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"gorm.io/gorm"
)

type IPluginService interface {
	CreatePlugin(ctx context.Context, userID, authorityID uint, req dto.CreatePluginReq) (*dto.PluginItem, error)
	UpdatePlugin(ctx context.Context, userID, authorityID uint, req dto.UpdatePluginReq) error
	GetPluginList(ctx context.Context, userID, authorityID uint, req dto.SearchPluginReq) ([]dto.PluginItem, int64, error)
	GetProjectDetail(ctx context.Context, userID, authorityID uint, req dto.GetProjectDetailReq) (*dto.ProjectDetail, error)
	CreateRelease(ctx context.Context, userID, authorityID uint, req dto.CreateReleaseReq) (*dto.PluginReleaseItem, error)
	UpdateRelease(ctx context.Context, userID, authorityID uint, req dto.UpdateReleaseReq) error
	GetReleaseDetail(ctx context.Context, userID, authorityID uint, req dto.GetReleaseDetailReq) (*dto.PluginReleaseItem, error)
	TransitionRelease(ctx context.Context, userID, authorityID uint, req dto.TransitionReleaseReq) (*dto.PluginReleaseItem, error)
	ClaimWorkOrder(ctx context.Context, userID, authorityID uint, req dto.ClaimWorkOrderReq) (*dto.PluginReleaseItem, error)
	ResetWorkOrder(ctx context.Context, userID, authorityID uint, req dto.ResetWorkOrderReq) (*dto.PluginReleaseItem, error)
	GetWorkOrderPool(ctx context.Context, userID, authorityID uint, req dto.SearchWorkOrderReq) ([]dto.WorkOrderItem, int64, error)
	GetProductList(ctx context.Context, req dto.SearchProductReq) ([]dto.ProductItem, int64, error)
	CreateProduct(ctx context.Context, authorityID uint, req dto.CreateProductReq) error
	UpdateProduct(ctx context.Context, authorityID uint, req dto.UpdateProductReq) error
	GetDepartmentList(ctx context.Context, req dto.SearchDepartmentReq) ([]dto.DepartmentItem, int64, error)
	CreateDepartment(ctx context.Context, authorityID uint, req dto.CreateDepartmentReq) error
	UpdateDepartment(ctx context.Context, authorityID uint, req dto.UpdateDepartmentReq) error
	GetPublishedPluginList(ctx context.Context, req dto.GetPublishedPluginListReq) ([]dto.PublishedPluginItem, int64, error)
	GetPublishedPluginDetail(ctx context.Context, id uint) (*dto.PublishedPluginDetail, error)
}

type PluginService struct {
	svcCtx *svc.ServiceContext
	repo   repository.IPluginRepository
}

func NewPluginService(svcCtx *svc.ServiceContext, repo repository.IPluginRepository) IPluginService {
	return &PluginService{svcCtx: svcCtx, repo: repo}
}

func (s *PluginService) CreatePlugin(ctx context.Context, userID, authorityID uint, req dto.CreatePluginReq) (*dto.PluginItem, error) {
	if !canManagePlugin(authorityID) {
		return nil, errcode.PluginForbidden
	}
	ownerID := req.OwnerID
	if ownerID == 0 {
		ownerID = userID
	}
	item := &model.Plugin{
		Code:          strings.TrimSpace(req.Code),
		RepositoryURL: strings.TrimSpace(req.RepositoryURL),
		NameZh:        strings.TrimSpace(req.NameZh),
		NameEn:        strings.TrimSpace(req.NameEn),
		DescriptionZh: strings.TrimSpace(req.DescriptionZh),
		DescriptionEn: strings.TrimSpace(req.DescriptionEn),
		DepartmentID:  req.DepartmentID,
		OwnerID:       ownerID,
		CreatedBy:     userID,
	}
	if err := s.repo.CreatePlugin(ctx, item); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindPluginByID(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	resp := toPluginItem(loaded)
	return &resp, nil
}

func (s *PluginService) UpdatePlugin(ctx context.Context, userID, authorityID uint, req dto.UpdatePluginReq) error {
	item, err := s.repo.FindPluginByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if !canEditPlugin(authorityID, userID, item) {
		return errcode.PluginForbidden
	}
	ownerID := req.OwnerID
	if ownerID == 0 {
		ownerID = item.OwnerID
	}
	return s.repo.UpdatePlugin(ctx, item, map[string]interface{}{
		"repository_url": strings.TrimSpace(req.RepositoryURL),
		"name_zh":        strings.TrimSpace(req.NameZh),
		"name_en":        strings.TrimSpace(req.NameEn),
		"description_zh": strings.TrimSpace(req.DescriptionZh),
		"description_en": strings.TrimSpace(req.DescriptionEn),
		"department_id":  req.DepartmentID,
		"owner_id":       ownerID,
	})
}

func (s *PluginService) GetPluginList(ctx context.Context, userID, authorityID uint, req dto.SearchPluginReq) ([]dto.PluginItem, int64, error) {
	query := s.repo.DB().Model(&model.Plugin{})
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+strings.TrimSpace(req.Code)+"%")
	}
	if req.Name != "" {
		keyword := "%" + strings.TrimSpace(req.Name) + "%"
		query = query.Where("name_zh LIKE ? OR name_en LIKE ?", keyword, keyword)
	}
	if !isAdmin(authorityID) && !isReviewer(authorityID) {
		query = query.Where("created_by = ? OR owner_id = ?", userID, userID)
	}
	items, total, err := s.repo.ListPlugins(ctx, query, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.PluginItem, 0, len(items))
	for _, item := range items {
		resp = append(resp, toPluginItem(&item))
	}
	return resp, total, nil
}

func (s *PluginService) GetProjectDetail(ctx context.Context, userID, authorityID uint, req dto.GetProjectDetailReq) (*dto.ProjectDetail, error) {
	item, err := s.repo.FindPluginByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if !canViewPlugin(authorityID, userID, item) {
		return nil, errcode.PluginForbidden
	}
	releases, err := s.repo.ListReleasesByPluginID(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	detail := &dto.ProjectDetail{
		Plugin:   toPluginItem(item),
		Releases: make([]dto.PluginReleaseItem, 0, len(releases)),
	}
	var releaseID uint
	if req.ReleaseID != nil {
		releaseID = *req.ReleaseID
	}
	for _, release := range releases {
		releaseCopy := release
		mapped := toReleaseItem(&releaseCopy)
		detail.Releases = append(detail.Releases, mapped)
		if releaseID == 0 || releaseID == release.ID {
			current := mapped
			detail.SelectedRelease = &current
			releaseID = release.ID
		}
	}
	if releaseID != 0 {
		events, err := s.repo.ListEventsByReleaseID(ctx, releaseID)
		if err != nil {
			return nil, err
		}
		detail.Events = toEventItems(events)
	}
	return detail, nil
}

func (s *PluginService) CreateRelease(ctx context.Context, userID, authorityID uint, req dto.CreateReleaseReq) (*dto.PluginReleaseItem, error) {
	item, err := s.repo.FindPluginByID(ctx, req.PluginID)
	if err != nil {
		return nil, err
	}
	if !canEditPlugin(authorityID, userID, item) {
		return nil, errcode.PluginForbidden
	}

	version, err := s.validateReleaseVersion(ctx, req.PluginID, 0, req.RequestType, req.Version)
	if err != nil {
		return nil, err
	}
	compatibles, universal, err := s.buildCompatibles(ctx, req.Compatibility)
	if err != nil {
		return nil, err
	}

	release := &model.PluginRelease{
		PluginID:        req.PluginID,
		RequestType:     req.RequestType,
		Status:          model.ReleaseStatusReady,
		ProcessStatus:   model.ReleaseProcessStatusDone,
		Version:         version,
		Universal:       universal,
		TestReportURL:   strings.TrimSpace(req.TestReportURL),
		PackageX86URL:   strings.TrimSpace(req.PackageX86URL),
		PackageARMURL:   strings.TrimSpace(req.PackageARMURL),
		ChangelogZh:     strings.TrimSpace(req.ChangelogZh),
		ChangelogEn:     strings.TrimSpace(req.ChangelogEn),
		OfflineReasonZh: strings.TrimSpace(req.OfflineReasonZh),
		OfflineReasonEn: strings.TrimSpace(req.OfflineReasonEn),
		TDID:            strings.TrimSpace(req.TDID),
		CreatedBy:       userID,
	}
	if err := s.repo.CreateRelease(ctx, release, compatibles); err != nil {
		return nil, err
	}
	_ = s.repo.CreateEvent(ctx, s.newEvent(release.ID, 0, release.Status, 0, release.ProcessStatus, model.ReleaseActionCreate, userID, "create release"))

	loaded, err := s.repo.FindReleaseByID(ctx, release.ID)
	if err != nil {
		return nil, err
	}
	resp := toReleaseItem(loaded)
	return &resp, nil
}

func (s *PluginService) UpdateRelease(ctx context.Context, userID, authorityID uint, req dto.UpdateReleaseReq) error {
	release, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if !canEditRelease(authorityID, userID, release) {
		if isProvider(authorityID) && ownsRelease(userID, release) {
			return errcode.PluginReleaseNotEditable
		}
		return errcode.PluginForbidden
	}

	version, err := s.validateReleaseVersion(ctx, release.PluginID, release.ID, release.RequestType, req.Version)
	if err != nil {
		return err
	}
	compatibles, universal, err := s.buildCompatibles(ctx, req.Compatibility)
	if err != nil {
		return err
	}

	return s.repo.UpdateRelease(ctx, release, map[string]interface{}{
		"version":           version,
		"universal":         universal,
		"test_report_url":   strings.TrimSpace(req.TestReportURL),
		"package_x86_url":   strings.TrimSpace(req.PackageX86URL),
		"package_arm_url":   strings.TrimSpace(req.PackageARMURL),
		"changelog_zh":      strings.TrimSpace(req.ChangelogZh),
		"changelog_en":      strings.TrimSpace(req.ChangelogEn),
		"offline_reason_zh": strings.TrimSpace(req.OfflineReasonZh),
		"offline_reason_en": strings.TrimSpace(req.OfflineReasonEn),
		"td_id":             strings.TrimSpace(req.TDID),
	}, compatibles)
}

func (s *PluginService) GetReleaseDetail(ctx context.Context, userID, authorityID uint, req dto.GetReleaseDetailReq) (*dto.PluginReleaseItem, error) {
	release, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if !canViewPlugin(authorityID, userID, &release.Plugin) {
		return nil, errcode.PluginForbidden
	}
	resp := toReleaseItem(release)
	return &resp, nil
}

func (s *PluginService) TransitionRelease(ctx context.Context, userID, authorityID uint, req dto.TransitionReleaseReq) (*dto.PluginReleaseItem, error) {
	release, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	updates := map[string]interface{}{}
	fromStatus := release.Status
	fromProcess := release.ProcessStatus

	switch req.Action {
	case model.ReleaseActionSubmitReview:
		if !canOperateOwnRelease(authorityID, userID, release) || release.Status != model.ReleaseStatusReady {
			return nil, errcode.PluginStatusInvalid
		}
		updates["status"] = model.ReleaseStatusPendingReview
		updates["process_status"] = model.ReleaseProcessStatusPending
		updates["submitted_at"] = now
		updates["claimer_id"] = nil
		updates["claimed_at"] = nil
	case model.ReleaseActionApprove:
		if !canReviewRelease(authorityID, userID, release) || release.Status != model.ReleaseStatusPendingReview || release.ProcessStatus != model.ReleaseProcessStatusProcessing {
			return nil, errcode.PluginStatusInvalid
		}
		updates["status"] = model.ReleaseStatusApproved
		updates["approved_at"] = now
		updates["review_comment"] = strings.TrimSpace(req.ReviewComment)
	case model.ReleaseActionReject:
		if strings.TrimSpace(req.ReviewComment) == "" {
			return nil, errcode.PluginReviewCommentRequired
		}
		if !canReviewRelease(authorityID, userID, release) || release.Status != model.ReleaseStatusPendingReview || release.ProcessStatus != model.ReleaseProcessStatusProcessing {
			return nil, errcode.PluginStatusInvalid
		}
		updates["status"] = model.ReleaseStatusRejected
		updates["process_status"] = model.ReleaseProcessStatusRejected
		updates["review_comment"] = strings.TrimSpace(req.ReviewComment)
	case model.ReleaseActionRevise:
		if !canOperateOwnRelease(authorityID, userID, release) || release.Status != model.ReleaseStatusRejected {
			return nil, errcode.PluginStatusInvalid
		}
		updates["status"] = model.ReleaseStatusReady
		updates["process_status"] = model.ReleaseProcessStatusDone
		updates["claimer_id"] = nil
		updates["claimed_at"] = nil
	case model.ReleaseActionRelease:
		if !canReviewRelease(authorityID, userID, release) || release.Status != model.ReleaseStatusApproved || release.ProcessStatus != model.ReleaseProcessStatusProcessing {
			return nil, errcode.PluginStatusInvalid
		}
		updates["status"] = model.ReleaseStatusReleased
		updates["process_status"] = model.ReleaseProcessStatusDone
		updates["released_at"] = now
	case model.ReleaseActionRequestOffline:
		if !canOperateOwnRelease(authorityID, userID, release) || release.Status != model.ReleaseStatusReleased {
			return nil, errcode.PluginStatusInvalid
		}
		updates["request_type"] = model.ReleaseRequestTypeOffline
		updates["status"] = model.ReleaseStatusPendingReview
		updates["process_status"] = model.ReleaseProcessStatusPending
		updates["td_id"] = strings.TrimSpace(req.TDID)
		updates["offline_reason_zh"] = strings.TrimSpace(req.OfflineReasonZh)
		updates["offline_reason_en"] = strings.TrimSpace(req.OfflineReasonEn)
		updates["submitted_at"] = now
		updates["claimer_id"] = nil
		updates["claimed_at"] = nil
	case model.ReleaseActionOffline:
		if !canReviewRelease(authorityID, userID, release) || release.RequestType != model.ReleaseRequestTypeOffline || release.Status != model.ReleaseStatusApproved || release.ProcessStatus != model.ReleaseProcessStatusProcessing {
			return nil, errcode.PluginStatusInvalid
		}
		updates["status"] = model.ReleaseStatusOfflined
		updates["process_status"] = model.ReleaseProcessStatusDone
		updates["offlined_at"] = now
	default:
		return nil, errcode.PluginStatusInvalid
	}

	if err := s.repo.UpdateRelease(ctx, release, updates, release.CompatibleItems); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindReleaseByID(ctx, release.ID)
	if err != nil {
		return nil, err
	}

	comment := strings.TrimSpace(req.ReviewComment)
	if comment == "" {
		comment = strings.TrimSpace(req.OfflineReasonZh)
	}
	_ = s.repo.CreateEvent(ctx, s.newEvent(release.ID, fromStatus, loaded.Status, fromProcess, loaded.ProcessStatus, req.Action, userID, comment))

	resp := toReleaseItem(loaded)
	return &resp, nil
}

func (s *PluginService) ClaimWorkOrder(ctx context.Context, userID, authorityID uint, req dto.ClaimWorkOrderReq) (*dto.PluginReleaseItem, error) {
	if !isReviewer(authorityID) {
		return nil, errcode.PluginForbidden
	}
	ok, err := s.repo.ClaimRelease(ctx, req.ID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errcode.PluginWorkOrderAlreadyClaim
	}
	loaded, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	_ = s.repo.CreateEvent(ctx, s.newEvent(req.ID, loaded.Status, loaded.Status, model.ReleaseProcessStatusPending, loaded.ProcessStatus, model.ReleaseActionClaim, userID, "claim work order"))
	resp := toReleaseItem(loaded)
	return &resp, nil
}

func (s *PluginService) ResetWorkOrder(ctx context.Context, userID, authorityID uint, req dto.ResetWorkOrderReq) (*dto.PluginReleaseItem, error) {
	if !isAdmin(authorityID) {
		return nil, errcode.PluginForbidden
	}
	if strings.TrimSpace(req.Reason) == "" {
		return nil, errcode.PluginResetReasonRequired
	}
	release, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if release.ProcessStatus != model.ReleaseProcessStatusProcessing {
		return nil, errcode.PluginStatusInvalid
	}
	if err := s.repo.ResetClaim(ctx, req.ID); err != nil {
		return nil, err
	}
	loaded, err := s.repo.FindReleaseByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	_ = s.repo.CreateEvent(ctx, s.newEvent(req.ID, loaded.Status, loaded.Status, model.ReleaseProcessStatusProcessing, loaded.ProcessStatus, model.ReleaseActionReset, userID, strings.TrimSpace(req.Reason)))
	resp := toReleaseItem(loaded)
	return &resp, nil
}

func (s *PluginService) GetWorkOrderPool(ctx context.Context, userID, authorityID uint, req dto.SearchWorkOrderReq) ([]dto.WorkOrderItem, int64, error) {
	if !isReviewer(authorityID) {
		return nil, 0, errcode.PluginForbidden
	}

	query := s.repo.DB().Model(&model.PluginRelease{})
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = dto.WorkOrderScopeAll
	}
	if scope == dto.WorkOrderScopeMine {
		query = query.Where("claimer_id = ?", userID)
	}
	if req.ProcessStatus != nil {
		query = query.Where("process_status = ?", *req.ProcessStatus)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.RequestType != nil {
		query = query.Where("request_type = ?", *req.RequestType)
	}
	if req.ClaimerID != nil {
		query = query.Where("claimer_id = ?", *req.ClaimerID)
	}
	if req.PluginID != nil {
		query = query.Where("plugin_id = ?", *req.PluginID)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Joins("JOIN plugins ON plugins.id = plugin_releases.plugin_id").
			Where("plugins.code LIKE ? OR plugins.name_zh LIKE ? OR plugins.name_en LIKE ? OR plugin_releases.version LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	items, total, err := s.repo.ListWorkOrders(ctx, query, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.WorkOrderItem, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.WorkOrderItem{PluginReleaseItem: toReleaseItem(&item)})
	}
	return resp, total, nil
}

func (s *PluginService) GetProductList(ctx context.Context, req dto.SearchProductReq) ([]dto.ProductItem, int64, error) {
	items, total, err := s.repo.ListProductsWithInactive(ctx, req.Page, req.PageSize, req.IncludeInactive)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.ProductItem, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.ProductItem{
			ID:          item.ID,
			Code:        item.Code,
			Name:        item.Name,
			Type:        item.Type,
			Description: item.Description,
			Status:      item.Status,
		})
	}
	return resp, total, nil
}

func (s *PluginService) CreateProduct(ctx context.Context, authorityID uint, req dto.CreateProductReq) error {
	if !isAdmin(authorityID) {
		return errcode.PluginForbidden
	}
	if !isCompatibleType(req.Type) {
		return errcode.PluginProductInvalid
	}
	return s.repo.CreateProduct(ctx, &model.PluginProduct{
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Type:        strings.TrimSpace(req.Type),
		Description: strings.TrimSpace(req.Description),
		Status:      true,
	})
}

func (s *PluginService) UpdateProduct(ctx context.Context, authorityID uint, req dto.UpdateProductReq) error {
	if !isAdmin(authorityID) {
		return errcode.PluginForbidden
	}
	if !isCompatibleType(req.Type) {
		return errcode.PluginProductInvalid
	}
	item, err := s.repo.FindProductByID(ctx, req.ID)
	if err != nil {
		return err
	}
	return s.repo.UpdateProduct(ctx, item, map[string]interface{}{
		"name":        strings.TrimSpace(req.Name),
		"type":        strings.TrimSpace(req.Type),
		"description": strings.TrimSpace(req.Description),
		"status":      req.Status,
	})
}

func (s *PluginService) GetDepartmentList(ctx context.Context, req dto.SearchDepartmentReq) ([]dto.DepartmentItem, int64, error) {
	items, total, err := s.repo.ListDepartmentsWithInactive(ctx, req.Page, req.PageSize, req.IncludeInactive)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.DepartmentItem, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.DepartmentItem{
			ID:          item.ID,
			Name:        firstNonEmpty(item.NameZh, item.Name),
			NameZh:      firstNonEmpty(item.NameZh, item.Name),
			NameEn:      firstNonEmpty(item.NameEn, item.Name),
			ProductLine: item.ProductLine,
			Status:      item.Status,
		})
	}
	return resp, total, nil
}

func (s *PluginService) CreateDepartment(ctx context.Context, authorityID uint, req dto.CreateDepartmentReq) error {
	if !isAdmin(authorityID) {
		return errcode.PluginForbidden
	}
	return s.repo.CreateDepartment(ctx, &model.PluginDepartment{
		Name:        strings.TrimSpace(req.NameZh),
		NameZh:      strings.TrimSpace(req.NameZh),
		NameEn:      strings.TrimSpace(req.NameEn),
		ProductLine: strings.TrimSpace(req.ProductLine),
		ParentID:    req.ParentID,
		Status:      true,
	})
}

func (s *PluginService) UpdateDepartment(ctx context.Context, authorityID uint, req dto.UpdateDepartmentReq) error {
	if !isAdmin(authorityID) {
		return errcode.PluginForbidden
	}
	item, err := s.repo.FindDepartmentByID(ctx, req.ID)
	if err != nil {
		return err
	}
	return s.repo.UpdateDepartment(ctx, item, map[string]interface{}{
		"name":         strings.TrimSpace(req.NameZh),
		"name_zh":      strings.TrimSpace(req.NameZh),
		"name_en":      strings.TrimSpace(req.NameEn),
		"product_line": strings.TrimSpace(req.ProductLine),
		"parent_id":    req.ParentID,
		"status":       req.Status,
	})
}

func (s *PluginService) GetPublishedPluginList(ctx context.Context, req dto.GetPublishedPluginListReq) ([]dto.PublishedPluginItem, int64, error) {
	query := s.repo.DB().Model(&model.Plugin{}).Where("id IN (?)",
		s.repo.DB().Model(&model.PluginRelease{}).Select("plugin_id").Where("status = ?", model.ReleaseStatusReleased),
	)
	plugins, total, err := s.repo.ListPublishedPlugins(ctx, query, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.PublishedPluginItem, 0, len(plugins))
	for _, item := range plugins {
		release, err := s.repo.FindLatestPublishedReleaseByPluginID(ctx, item.ID)
		if err != nil {
			continue
		}
		resp = append(resp, dto.PublishedPluginItem{
			ID:              item.ID,
			Code:            item.Code,
			NameZh:          item.NameZh,
			NameEn:          item.NameEn,
			DescriptionZh:   item.DescriptionZh,
			DescriptionEn:   item.DescriptionEn,
			LatestVersion:   release.Version,
			CompatibleItems: toCompatibleItems(release.CompatibleItems),
		})
	}
	return resp, total, nil
}

func (s *PluginService) GetPublishedPluginDetail(ctx context.Context, id uint) (*dto.PublishedPluginDetail, error) {
	plugin, release, err := s.repo.FindPublishedPluginByID(ctx, id)
	if err != nil {
		return nil, err
	}
	versions, err := s.repo.ListPublishedReleasesByPluginID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := &dto.PublishedPluginDetail{
		Plugin:   toPluginItem(plugin),
		Release:  toPublishedReleaseItem(release),
		Versions: make([]dto.PublishedReleaseItem, 0, len(versions)),
	}
	for _, item := range versions {
		releaseCopy := item
		resp.Versions = append(resp.Versions, toPublishedReleaseItem(&releaseCopy))
	}
	return resp, nil
}

func (s *PluginService) validateReleaseVersion(ctx context.Context, pluginID, releaseID uint, requestType int8, rawVersion string) (string, error) {
	if requestType != model.ReleaseRequestTypeVersion {
		return strings.TrimSpace(rawVersion), nil
	}
	version, ok := normalizeSemver(strings.TrimSpace(rawVersion))
	if !ok {
		return "", errcode.PluginVersionInvalid
	}

	dup, err := s.repo.FindReleaseByPluginVersion(ctx, pluginID, version)
	if err == nil && dup.ID != releaseID {
		return "", errcode.PluginVersionDuplicate
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	highest, err := s.repo.FindHighestVersionReleaseByPluginID(ctx, pluginID, releaseID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if version != "1.0.0" {
			return "", errcode.PluginVersionInvalid
		}
		return version, nil
	}
	if err != nil {
		return "", err
	}
	if compareSemver(version, highest.Version) <= 0 {
		return "", errcode.PluginVersionInvalid
	}
	return version, nil
}

func (s *PluginService) buildCompatibles(ctx context.Context, compatibility dto.ReleaseCompatibilityReq) ([]model.PluginCompatibleProduct, bool, error) {
	if !compatibility.Universal && len(compatibility.ProductItems) == 0 && len(compatibility.AcliItems) == 0 {
		return nil, false, errcode.PluginCompatibilityRequired
	}

	result := make([]model.PluginCompatibleProduct, 0, len(compatibility.ProductItems)+len(compatibility.AcliItems))
	seen := map[uint]struct{}{}

	appendItems := func(items []dto.UpsertCompatibleProductReq, wantType string) error {
		for _, item := range items {
			if item.ProductID == 0 {
				return errcode.PluginProductInvalid
			}
			product, err := s.repo.FindProductByID(ctx, item.ProductID)
			if err != nil {
				return errcode.PluginProductInvalid
			}
			if product.Type != wantType {
				return errcode.PluginProductInvalid
			}
			if _, ok := seen[item.ProductID]; ok {
				continue
			}
			seen[item.ProductID] = struct{}{}
			result = append(result, model.PluginCompatibleProduct{
				ProductID:         item.ProductID,
				VersionConstraint: strings.TrimSpace(item.VersionConstraint),
			})
		}
		return nil
	}

	if err := appendItems(compatibility.ProductItems, model.CompatibleTargetTypeProduct); err != nil {
		return nil, false, err
	}
	if err := appendItems(compatibility.AcliItems, model.CompatibleTargetTypeAcli); err != nil {
		return nil, false, err
	}
	return result, compatibility.Universal, nil
}

func (s *PluginService) newEvent(releaseID uint, fromStatus, toStatus, fromProcess, toProcess int8, action string, operatorID uint, comment string) *model.PluginReleaseEvent {
	snapshot, _ := json.Marshal(map[string]interface{}{
		"releaseId":  releaseID,
		"action":     action,
		"operatorId": operatorID,
		"comment":    comment,
	})
	return &model.PluginReleaseEvent{
		ReleaseID:         releaseID,
		FromStatus:        fromStatus,
		ToStatus:          toStatus,
		FromProcessStatus: fromProcess,
		ToProcessStatus:   toProcess,
		Action:            action,
		OperatorID:        operatorID,
		Comment:           comment,
		SnapshotJSON:      string(snapshot),
	}
}

func isAdmin(authorityID uint) bool    { return authorityID == 1 }
func isReviewer(authorityID uint) bool { return authorityID == 1 || authorityID == 10013 }
func isProvider(authorityID uint) bool { return authorityID == 1 || authorityID == 10010 }

func canManagePlugin(authorityID uint) bool { return isProvider(authorityID) }

func canViewPlugin(authorityID, userID uint, item *model.Plugin) bool {
	if isReviewer(authorityID) {
		return true
	}
	return item.OwnerID == userID || item.CreatedBy == userID || isAdmin(authorityID)
}

func canEditPlugin(authorityID, userID uint, item *model.Plugin) bool {
	return isAdmin(authorityID) || (isProvider(authorityID) && (item.OwnerID == userID || item.CreatedBy == userID))
}

func ownsRelease(userID uint, release *model.PluginRelease) bool {
	return release.CreatedBy == userID || release.Plugin.OwnerID == userID || release.Plugin.CreatedBy == userID
}

func canOperateOwnRelease(authorityID, userID uint, release *model.PluginRelease) bool {
	return isProvider(authorityID) && ownsRelease(userID, release)
}

func canEditRelease(authorityID, userID uint, release *model.PluginRelease) bool {
	if !canOperateOwnRelease(authorityID, userID, release) {
		return false
	}
	return release.Status == model.ReleaseStatusReady || release.Status == model.ReleaseStatusRejected
}

func canReviewRelease(authorityID, userID uint, release *model.PluginRelease) bool {
	if isAdmin(authorityID) {
		return true
	}
	return isReviewer(authorityID) && release.ClaimerID != nil && *release.ClaimerID == userID
}

func toPluginItem(item *model.Plugin) dto.PluginItem {
	department := ""
	if item.Department.ID != 0 {
		department = firstNonEmpty(item.Department.NameZh, item.Department.Name)
	}
	return dto.PluginItem{
		ID:            item.ID,
		Code:          item.Code,
		RepositoryURL: item.RepositoryURL,
		NameZh:        item.NameZh,
		NameEn:        item.NameEn,
		DescriptionZh: item.DescriptionZh,
		DescriptionEn: item.DescriptionEn,
		DepartmentID:  item.DepartmentID,
		Department:    department,
		OwnerID:       item.OwnerID,
		CreatedBy:     item.CreatedBy,
		CreatedAt:     item.CreatedAt.Format(time.RFC3339),
	}
}

func toReleaseItem(item *model.PluginRelease) dto.PluginReleaseItem {
	compatibility := toReleaseCompatibility(item)
	resp := dto.PluginReleaseItem{
		ID:              item.ID,
		PluginID:        item.PluginID,
		PluginCode:      item.Plugin.Code,
		PluginNameZh:    item.Plugin.NameZh,
		RequestType:     item.RequestType,
		Status:          item.Status,
		ProcessStatus:   item.ProcessStatus,
		Version:         item.Version,
		ClaimerID:       item.ClaimerID,
		ClaimerName:     strings.TrimSpace(item.Claimer.NickName),
		ClaimerUsername: strings.TrimSpace(item.Claimer.Username),
		ReviewComment:   item.ReviewComment,
		TestReportURL:   item.TestReportURL,
		PackageX86URL:   item.PackageX86URL,
		PackageARMURL:   item.PackageARMURL,
		ChangelogZh:     item.ChangelogZh,
		ChangelogEn:     item.ChangelogEn,
		OfflineReasonZh: item.OfflineReasonZh,
		OfflineReasonEn: item.OfflineReasonEn,
		TDID:            item.TDID,
		Compatibility:   compatibility,
		CompatibleItems: append(append([]dto.CompatibleProductItem{}, compatibility.ProductItems...), compatibility.AcliItems...),
		CreatedBy:       item.CreatedBy,
		CreatedAt:       item.CreatedAt.Format(time.RFC3339),
	}
	if item.SubmittedAt != nil {
		v := item.SubmittedAt.Format(time.RFC3339)
		resp.SubmittedAt = &v
	}
	if item.ApprovedAt != nil {
		v := item.ApprovedAt.Format(time.RFC3339)
		resp.ApprovedAt = &v
	}
	if item.ReleasedAt != nil {
		v := item.ReleasedAt.Format(time.RFC3339)
		resp.ReleasedAt = &v
	}
	if item.OfflinedAt != nil {
		v := item.OfflinedAt.Format(time.RFC3339)
		resp.OfflinedAt = &v
	}
	if item.ClaimedAt != nil {
		v := item.ClaimedAt.Format(time.RFC3339)
		resp.ClaimedAt = &v
	}
	return resp
}

func toReleaseCompatibility(item *model.PluginRelease) dto.ReleaseCompatibility {
	result := dto.ReleaseCompatibility{
		ProductItems: make([]dto.CompatibleProductItem, 0),
		AcliItems:    make([]dto.CompatibleProductItem, 0),
		Universal:    item.Universal,
	}
	for _, compatible := range item.CompatibleItems {
		mapped := dto.CompatibleProductItem{
			ID:                compatible.ID,
			ProductID:         compatible.ProductID,
			ProductCode:       compatible.Product.Code,
			ProductName:       compatible.Product.Name,
			Type:              compatible.Product.Type,
			VersionConstraint: compatible.VersionConstraint,
		}
		switch compatible.Product.Type {
		case model.CompatibleTargetTypeAcli:
			result.AcliItems = append(result.AcliItems, mapped)
		default:
			result.ProductItems = append(result.ProductItems, mapped)
		}
	}
	return result
}

func toCompatibleItems(items []model.PluginCompatibleProduct) []dto.CompatibleProductItem {
	resp := make([]dto.CompatibleProductItem, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.CompatibleProductItem{
			ID:                item.ID,
			ProductID:         item.ProductID,
			ProductCode:       item.Product.Code,
			ProductName:       item.Product.Name,
			Type:              item.Product.Type,
			VersionConstraint: item.VersionConstraint,
		})
	}
	return resp
}

func toEventItems(items []model.PluginReleaseEvent) []dto.EventItem {
	resp := make([]dto.EventItem, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.EventItem{
			ID:                item.ID,
			Action:            item.Action,
			FromStatus:        item.FromStatus,
			ToStatus:          item.ToStatus,
			FromProcessStatus: item.FromProcessStatus,
			ToProcessStatus:   item.ToProcessStatus,
			OperatorID:        item.OperatorID,
			Comment:           item.Comment,
			CreatedAt:         item.CreatedAt.Format(time.RFC3339),
		})
	}
	return resp
}

func toPublishedReleaseItem(item *model.PluginRelease) dto.PublishedReleaseItem {
	resp := dto.PublishedReleaseItem{
		ID:              item.ID,
		Version:         item.Version,
		ChangelogZh:     item.ChangelogZh,
		ChangelogEn:     item.ChangelogEn,
		TestReportURL:   item.TestReportURL,
		PackageX86URL:   item.PackageX86URL,
		PackageARMURL:   item.PackageARMURL,
		CompatibleItems: toCompatibleItems(item.CompatibleItems),
	}
	if item.ReleasedAt != nil {
		v := item.ReleasedAt.Format(time.RFC3339)
		resp.ReleasedAt = &v
	}
	return resp
}

func normalizeSemver(version string) (string, bool) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", false
	}
	normalized := make([]string, 0, 3)
	for _, part := range parts {
		if part == "" {
			return "", false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return "", false
		}
		normalized = append(normalized, strconv.Itoa(n))
	}
	return strings.Join(normalized, "."), true
}

func compareSemver(left, right string) int {
	lv, lok := normalizeSemver(left)
	rv, rok := normalizeSemver(right)
	if !lok || !rok {
		return strings.Compare(left, right)
	}
	lparts := strings.Split(lv, ".")
	rparts := strings.Split(rv, ".")
	for i := 0; i < 3; i++ {
		li, _ := strconv.Atoi(lparts[i])
		ri, _ := strconv.Atoi(rparts[i])
		if li < ri {
			return -1
		}
		if li > ri {
			return 1
		}
	}
	return 0
}

func isCompatibleType(value string) bool {
	switch strings.TrimSpace(value) {
	case model.CompatibleTargetTypeProduct, model.CompatibleTargetTypeAcli:
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
