package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/dto"
	pluginModel "github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func commonPageAll() common.PageInfo {
	return common.PageInfo{Page: 1, PageSize: 500}
}

type IPluginRepository interface {
	ListPlugins(ctx context.Context, req dto.SearchPluginReq) ([]dto.PluginListItem, int64, error)
	GetPluginOverview(ctx context.Context) (*dto.PluginOverview, error)
	ListPublishedPlugins(ctx context.Context, req dto.SearchPublishedPluginReq) ([]dto.PublishedPluginListItem, int64, error)
	GetPublishedPluginDetail(ctx context.Context, pluginID uint) (*dto.PublishedPluginDetail, error)
	GetProjectDetail(ctx context.Context, id uint) (*dto.ProjectDetail, error)
	CreatePlugin(ctx context.Context, plugin *pluginModel.Plugin) error
	UpdatePlugin(ctx context.Context, plugin *pluginModel.Plugin) error
	FindPluginByID(ctx context.Context, id uint) (*pluginModel.Plugin, error)
	FindPluginByCode(ctx context.Context, code string, excludeID uint) (*pluginModel.Plugin, error)
	FindPluginByRepo(ctx context.Context, repo string, excludeID uint) (*pluginModel.Plugin, error)

	ListReleases(ctx context.Context, req dto.SearchReleaseReq) ([]dto.ReleaseListItem, int64, error)
	GetReleaseDetail(ctx context.Context, id uint) (*dto.ReleaseDetail, error)
	CreateRelease(ctx context.Context, release *pluginModel.PluginRelease) error
	UpdateRelease(ctx context.Context, release *pluginModel.PluginRelease) error
	FindReleaseByID(ctx context.Context, id uint) (*pluginModel.PluginRelease, error)
	FindReleaseByVersion(ctx context.Context, pluginID uint, version string, excludeID uint) (*pluginModel.PluginRelease, error)
	CountActiveReleasedVersions(ctx context.Context, pluginID uint) (int64, error)
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type PluginRepository struct {
	db *gorm.DB
}

func NewPluginRepository(db *gorm.DB) IPluginRepository {
	return &PluginRepository{db: db}
}

func (r *PluginRepository) ListPlugins(ctx context.Context, req dto.SearchPluginReq) ([]dto.PluginListItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	base := r.db.WithContext(ctx).Model(&pluginModel.Plugin{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		base = base.Where("code LIKE ? OR name_zh LIKE ? OR name_en LIKE ? OR owner LIKE ?", like, like, like, like)
	}
	if req.CurrentStatus != "" {
		base = base.Where("current_status = ?", req.CurrentStatus)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		pluginModel.Plugin
		ReleaseCount int64 `gorm:"column:release_count"`
	}
	var rows []row
	err := base.Select("plugins.*, COUNT(pr.id) AS release_count").
		Joins("LEFT JOIN plugin_releases pr ON pr.plugin_id = plugins.id").
		Group("plugins.id").
		Order("plugins.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	list := make([]dto.PluginListItem, 0, len(rows))
	for _, item := range rows {
		list = append(list, dto.PluginListItem{
			ID:             item.ID,
			Code:           item.Code,
			RepositoryURL:  item.RepositoryURL,
			NameZh:         item.NameZh,
			NameEn:         item.NameEn,
			DescriptionZh:  item.DescriptionZh,
			DescriptionEn:  item.DescriptionEn,
			CapabilityZh:   item.CapabilityZh,
			CapabilityEn:   item.CapabilityEn,
			Owner:          item.Owner,
			CurrentStatus:  item.CurrentStatus,
			LatestVersion:  item.LatestVersion,
			LastReleasedAt: formatTime(item.LastReleasedAt),
			ReleaseCount:   item.ReleaseCount,
		})
	}
	if err := r.attachPluginSummaries(ctx, list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *PluginRepository) attachPluginSummaries(ctx context.Context, list []dto.PluginListItem) error {
	if len(list) == 0 {
		return nil
	}

	pluginIDs := make([]uint, 0, len(list))
	for _, item := range list {
		pluginIDs = append(pluginIDs, item.ID)
	}

	var releases []pluginModel.PluginRelease
	if err := r.db.WithContext(ctx).
		Where("plugin_id IN ?", pluginIDs).
		Order("created_at DESC, id DESC").
		Find(&releases).Error; err != nil {
		return err
	}

	summaryMap := make(map[uint]*dto.PluginListItem, len(list))
	for index := range list {
		summaryMap[list[index].ID] = &list[index]
	}

	for _, item := range releases {
		target := summaryMap[item.PluginID]
		if target == nil {
			continue
		}

		switch item.Status {
		case pluginModel.PluginReleaseStatusDraft, pluginModel.PluginReleaseStatusReleasePreparing:
			target.PreparingCount++
		case pluginModel.PluginReleaseStatusPendingReview:
			target.PendingReviewCount++
		case pluginModel.PluginReleaseStatusApproved:
			target.ApprovedCount++
		case pluginModel.PluginReleaseStatusReleased:
			if !item.IsOfflined {
				target.PublishedCount++
			}
		case pluginModel.PluginReleaseStatusOfflined:
			target.OfflinedCount++
		}

		if target.CurrentWorkflowID == nil {
			switch item.Status {
			case pluginModel.PluginReleaseStatusPendingReview,
				pluginModel.PluginReleaseStatusApproved,
				pluginModel.PluginReleaseStatusRejected,
				pluginModel.PluginReleaseStatusReleasePreparing,
				pluginModel.PluginReleaseStatusDraft:
				releaseID := item.ID
				target.CurrentWorkflowID = &releaseID
				target.CurrentWorkflowType = item.RequestType
				target.CurrentWorkflowStatus = item.Status
				target.CurrentWorkflowVersion = item.Version
			}
		}
	}

	return nil
}

func (r *PluginRepository) GetPluginOverview(ctx context.Context) (*dto.PluginOverview, error) {
	overview := &dto.PluginOverview{}
	if err := r.db.WithContext(ctx).Model(&pluginModel.Plugin{}).Count(&overview.ProjectCount).Error; err != nil {
		return nil, err
	}

	var releases []pluginModel.PluginRelease
	if err := r.db.WithContext(ctx).
		Select("status, is_offlined").
		Find(&releases).Error; err != nil {
		return nil, err
	}

	for _, item := range releases {
		switch item.Status {
		case pluginModel.PluginReleaseStatusDraft, pluginModel.PluginReleaseStatusReleasePreparing:
			overview.PreparingCount++
		case pluginModel.PluginReleaseStatusPendingReview:
			overview.PendingReviewCount++
		case pluginModel.PluginReleaseStatusApproved:
			overview.ApprovedCount++
		case pluginModel.PluginReleaseStatusReleased:
			if !item.IsOfflined {
				overview.PublishedCount++
			}
		case pluginModel.PluginReleaseStatusOfflined:
			overview.OfflinedCount++
		}
	}

	return overview, nil
}

func (r *PluginRepository) CreatePlugin(ctx context.Context, plugin *pluginModel.Plugin) error {
	return r.db.WithContext(ctx).Create(plugin).Error
}

func (r *PluginRepository) GetProjectDetail(ctx context.Context, id uint) (*dto.ProjectDetail, error) {
	plugin, err := r.FindPluginByID(ctx, id)
	if err != nil {
		return nil, err
	}

	releases, _, err := r.ListReleases(ctx, dto.SearchReleaseReq{
		PageInfo: commonPageAll(),
		PluginID: id,
	})
	if err != nil {
		return nil, err
	}

	detail := &dto.ProjectDetail{
		PluginListItem: dto.PluginListItem{
			ID:             plugin.ID,
			Code:           plugin.Code,
			RepositoryURL:  plugin.RepositoryURL,
			NameZh:         plugin.NameZh,
			NameEn:         plugin.NameEn,
			DescriptionZh:  plugin.DescriptionZh,
			DescriptionEn:  plugin.DescriptionEn,
			CapabilityZh:   plugin.CapabilityZh,
			CapabilityEn:   plugin.CapabilityEn,
			Owner:          plugin.Owner,
			CurrentStatus:  plugin.CurrentStatus,
			LatestVersion:  plugin.LatestVersion,
			LastReleasedAt: formatTime(plugin.LastReleasedAt),
			ReleaseCount:   int64(len(releases)),
		},
		Versions: releases,
	}

	for _, item := range releases {
		row := item
		switch item.Status {
		case pluginModel.PluginReleaseStatusDraft, pluginModel.PluginReleaseStatusReleasePreparing:
			detail.PreparingCount++
		case pluginModel.PluginReleaseStatusPendingReview:
			detail.PendingReviewCount++
		case pluginModel.PluginReleaseStatusApproved:
			detail.ApprovedCount++
		case pluginModel.PluginReleaseStatusReleased:
			if !item.IsOfflined {
				detail.PublishedCount++
			}
		case pluginModel.PluginReleaseStatusOfflined:
			detail.OfflinedCount++
		}

		if detail.CurrentWorkflow == nil {
			switch item.Status {
			case pluginModel.PluginReleaseStatusPendingReview,
				pluginModel.PluginReleaseStatusApproved,
				pluginModel.PluginReleaseStatusRejected,
				pluginModel.PluginReleaseStatusReleasePreparing,
				pluginModel.PluginReleaseStatusDraft:
				detail.CurrentWorkflow = &row
			}
		}

		if detail.LatestReleased == nil && item.Status == pluginModel.PluginReleaseStatusReleased && !item.IsOfflined {
			detail.LatestReleased = &row
		}
	}

	return detail, nil
}

func (r *PluginRepository) ListPublishedPlugins(ctx context.Context, req dto.SearchPublishedPluginReq) ([]dto.PublishedPluginListItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 12
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	base := r.db.WithContext(ctx).Model(&pluginModel.PluginRelease{}).
		Joins("JOIN plugins ON plugins.id = plugin_releases.plugin_id").
		Where("plugin_releases.status = ? AND plugin_releases.is_offlined = 0", pluginModel.PluginReleaseStatusReleased)
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		base = base.Where("plugins.code LIKE ? OR plugins.name_zh LIKE ? OR plugins.name_en LIKE ? OR plugin_releases.version LIKE ?", like, like, like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []pluginModel.PluginRelease
	err := base.Preload("Plugin").
		Order("plugin_releases.released_at DESC, plugin_releases.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	list := make([]dto.PublishedPluginListItem, 0, len(rows))
	for _, item := range rows {
		list = append(list, dto.PublishedPluginListItem{
			PluginID:             item.PluginID,
			ReleaseID:            item.ID,
			Code:                 item.Plugin.Code,
			NameZh:               item.Plugin.NameZh,
			NameEn:               item.Plugin.NameEn,
			DescriptionZh:        item.Plugin.DescriptionZh,
			DescriptionEn:        item.Plugin.DescriptionEn,
			CapabilityZh:         item.Plugin.CapabilityZh,
			CapabilityEn:         item.Plugin.CapabilityEn,
			Owner:                item.Plugin.Owner,
			Version:              item.Version,
			VersionConstraint:    item.VersionConstraint,
			Publisher:            item.Publisher,
			PackageX86URL:        item.PackageX86URL,
			PackageArmURL:        item.PackageArmURL,
			TestReportURL:        item.TestReportURL,
			ChangelogZh:          item.ChangelogZh,
			ChangelogEn:          item.ChangelogEn,
			PerformanceSummaryZh: item.PerformanceSummaryZh,
			PerformanceSummaryEn: item.PerformanceSummaryEn,
			ReleasedAt:           item.ReleasedAt.Format(time.RFC3339),
		})
	}

	return list, total, nil
}

func (r *PluginRepository) GetPublishedPluginDetail(ctx context.Context, pluginID uint) (*dto.PublishedPluginDetail, error) {
	var plugin pluginModel.Plugin
	if err := r.db.WithContext(ctx).First(&plugin, pluginID).Error; err != nil {
		return nil, err
	}

	var releases []pluginModel.PluginRelease
	if err := r.db.WithContext(ctx).
		Preload("Plugin").
		Where("plugin_id = ? AND status = ? AND is_offlined = 0", pluginID, pluginModel.PluginReleaseStatusReleased).
		Order("released_at DESC, id DESC").
		Find(&releases).Error; err != nil {
		return nil, err
	}

	detail := &dto.PublishedPluginDetail{
		PluginID:      plugin.ID,
		Code:          plugin.Code,
		NameZh:        plugin.NameZh,
		NameEn:        plugin.NameEn,
		DescriptionZh: plugin.DescriptionZh,
		DescriptionEn: plugin.DescriptionEn,
		CapabilityZh:  plugin.CapabilityZh,
		CapabilityEn:  plugin.CapabilityEn,
		Owner:         plugin.Owner,
		Versions:      make([]dto.PublishedPluginVersionItem, 0, len(releases)),
	}

	for _, item := range releases {
		releasedAt := ""
		if item.ReleasedAt != nil {
			releasedAt = item.ReleasedAt.Format(time.RFC3339)
		}
		detail.Versions = append(detail.Versions, dto.PublishedPluginVersionItem{
			ReleaseID:            item.ID,
			Version:              item.Version,
			VersionConstraint:    item.VersionConstraint,
			Publisher:            item.Publisher,
			PackageX86URL:        item.PackageX86URL,
			PackageArmURL:        item.PackageArmURL,
			TestReportURL:        item.TestReportURL,
			ChangelogZh:          item.ChangelogZh,
			ChangelogEn:          item.ChangelogEn,
			PerformanceSummaryZh: item.PerformanceSummaryZh,
			PerformanceSummaryEn: item.PerformanceSummaryEn,
			ReleasedAt:           releasedAt,
		})
	}

	return detail, nil
}

func (r *PluginRepository) UpdatePlugin(ctx context.Context, plugin *pluginModel.Plugin) error {
	return r.db.WithContext(ctx).Save(plugin).Error
}

func (r *PluginRepository) FindPluginByID(ctx context.Context, id uint) (*pluginModel.Plugin, error) {
	var plugin pluginModel.Plugin
	if err := r.db.WithContext(ctx).First(&plugin, id).Error; err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *PluginRepository) FindPluginByCode(ctx context.Context, code string, excludeID uint) (*pluginModel.Plugin, error) {
	var plugin pluginModel.Plugin
	query := r.db.WithContext(ctx).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&plugin).Error; err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *PluginRepository) FindPluginByRepo(ctx context.Context, repo string, excludeID uint) (*pluginModel.Plugin, error) {
	var plugin pluginModel.Plugin
	query := r.db.WithContext(ctx).Where("repository_url = ?", repo)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&plugin).Error; err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *PluginRepository) ListReleases(ctx context.Context, req dto.SearchReleaseReq) ([]dto.ReleaseListItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	base := r.db.WithContext(ctx).Model(&pluginModel.PluginRelease{}).
		Joins("JOIN plugins ON plugins.id = plugin_releases.plugin_id")
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		base = base.Where("plugins.code LIKE ? OR plugins.name_zh LIKE ? OR plugin_releases.version LIKE ?", like, like, like)
	}
	if req.PluginID > 0 {
		base = base.Where("plugin_releases.plugin_id = ?", req.PluginID)
	}
	if req.RequestType != "" {
		base = base.Where("plugin_releases.request_type = ?", req.RequestType)
	}
	if req.Status != "" {
		base = base.Where("plugin_releases.status = ?", req.Status)
	}
	if req.CreatedBy > 0 {
		base = base.Where("plugin_releases.created_by = ?", req.CreatedBy)
	}
	if req.ReviewerID > 0 {
		base = base.Where("plugin_releases.reviewer_id = ?", req.ReviewerID)
	}
	if req.PublisherID > 0 {
		base = base.Where("plugin_releases.publisher_id = ?", req.PublisherID)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []pluginModel.PluginRelease
	err := base.Preload("Plugin").
		Order("plugin_releases.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	list := make([]dto.ReleaseListItem, 0, len(rows))
	for _, item := range rows {
		list = append(list, buildReleaseListItem(item))
	}
	return list, total, nil
}

func (r *PluginRepository) GetReleaseDetail(ctx context.Context, id uint) (*dto.ReleaseDetail, error) {
	release, err := r.FindReleaseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	var events []pluginModel.PluginReleaseEvent
	if err := r.db.WithContext(ctx).
		Where("release_id = ?", id).
		Order("id ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	result := &dto.ReleaseDetail{
		ReleaseListItem: buildReleaseListItem(*release),
		Events:          make([]dto.ReleaseEventItem, 0, len(events)),
	}
	for _, item := range events {
		result.Events = append(result.Events, dto.ReleaseEventItem{
			ID:         item.ID,
			ReleaseID:  item.ReleaseID,
			FromStatus: string(item.FromStatus),
			ToStatus:   string(item.ToStatus),
			Action:     item.Action,
			OperatorID: item.OperatorID,
			Comment:    item.Comment,
			CreatedAt:  item.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (r *PluginRepository) CreateRelease(ctx context.Context, release *pluginModel.PluginRelease) error {
	return r.db.WithContext(ctx).Create(release).Error
}

func (r *PluginRepository) UpdateRelease(ctx context.Context, release *pluginModel.PluginRelease) error {
	return r.db.WithContext(ctx).Save(release).Error
}

func (r *PluginRepository) FindReleaseByID(ctx context.Context, id uint) (*pluginModel.PluginRelease, error) {
	var release pluginModel.PluginRelease
	if err := r.db.WithContext(ctx).Preload("Plugin").First(&release, id).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *PluginRepository) FindReleaseByVersion(ctx context.Context, pluginID uint, version string, excludeID uint) (*pluginModel.PluginRelease, error) {
	var release pluginModel.PluginRelease
	query := r.db.WithContext(ctx).Where("plugin_id = ? AND version = ?", pluginID, version)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&release).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *PluginRepository) CountActiveReleasedVersions(ctx context.Context, pluginID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&pluginModel.PluginRelease{}).
		Where("plugin_id = ? AND status = ? AND is_offlined = 0", pluginID, pluginModel.PluginReleaseStatusReleased).
		Count(&total).Error
	return total, err
}

func (r *PluginRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func buildReleaseListItem(item pluginModel.PluginRelease) dto.ReleaseListItem {
	var checklist []dto.PluginChecklistItem
	if len(item.Checklist) > 0 {
		_ = json.Unmarshal(item.Checklist, &checklist)
	}
	return dto.ReleaseListItem{
		ID:                   item.ID,
		PluginID:             item.PluginID,
		PluginCode:           item.Plugin.Code,
		PluginNameZh:         item.Plugin.NameZh,
		PluginNameEn:         item.Plugin.NameEn,
		RequestType:          item.RequestType,
		Status:               item.Status,
		Version:              item.Version,
		VersionConstraint:    item.VersionConstraint,
		Publisher:            item.Publisher,
		ReviewerID:           item.ReviewerID,
		PublisherID:          item.PublisherID,
		Checklist:            checklist,
		PerformanceSummaryZh: item.PerformanceSummaryZh,
		PerformanceSummaryEn: item.PerformanceSummaryEn,
		TestReportURL:        item.TestReportURL,
		PackageX86URL:        item.PackageX86URL,
		PackageArmURL:        item.PackageArmURL,
		ChangelogZh:          item.ChangelogZh,
		ChangelogEn:          item.ChangelogEn,
		ReviewComment:        item.ReviewComment,
		OfflineReasonZh:      item.OfflineReasonZh,
		OfflineReasonEn:      item.OfflineReasonEn,
		IsOfflined:           item.IsOfflined,
		SourceReleaseID:      item.SourceReleaseID,
		TargetReleaseID:      item.TargetReleaseID,
		CreatedBy:            item.CreatedBy,
		CreatedAt:            item.CreatedAt.Format(time.RFC3339),
		SubmittedAt:          formatTime(item.SubmittedAt),
		ApprovedAt:           formatTime(item.ApprovedAt),
		ReleasedAt:           formatTime(item.ReleasedAt),
		OfflinedAt:           formatTime(item.OfflinedAt),
	}
}

func formatTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}

func MarshalChecklist(items []dto.PluginChecklistItem) (datatypes.JSON, error) {
	if len(items) == 0 {
		return datatypes.JSON([]byte("[]")), nil
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(raw), nil
}
