package service

import (
	"context"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/model"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/repository"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AutoMigrate() error {
	return s.repo.AutoMigrate()
}

func (s *Service) UpsertPlugin(ctx context.Context, req dto.UpsertPluginReq) error {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "published"
	}
	return s.repo.UpsertPlugin(ctx, &model.MarketPlugin{
		PluginID:      req.PluginID,
		Code:          strings.TrimSpace(req.Code),
		NameZh:        strings.TrimSpace(req.NameZh),
		NameEn:        strings.TrimSpace(req.NameEn),
		DescriptionZh: strings.TrimSpace(req.DescriptionZh),
		DescriptionEn: strings.TrimSpace(req.DescriptionEn),
		CapabilityZh:  strings.TrimSpace(req.CapabilityZh),
		CapabilityEn:  strings.TrimSpace(req.CapabilityEn),
		OwnerName:     strings.TrimSpace(req.OwnerName),
		Status:        status,
	})
}

func (s *Service) UpsertVersion(ctx context.Context, req dto.UpsertVersionReq) error {
	plugin, err := s.repo.FindPluginByPluginID(ctx, req.PluginID)
	if err != nil {
		return err
	}

	var releasedAt *time.Time
	if trimmed := strings.TrimSpace(req.ReleasedAt); trimmed != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, trimmed); parseErr == nil {
			releasedAt = &parsed
		}
	}

	version := &model.MarketPluginVersion{
		PluginRefID:       plugin.ID,
		PluginID:          req.PluginID,
		ReleaseID:         req.ReleaseID,
		Version:           strings.TrimSpace(req.Version),
		ChangelogZh:       strings.TrimSpace(req.ChangelogZh),
		ChangelogEn:       strings.TrimSpace(req.ChangelogEn),
		TestReportURL:     strings.TrimSpace(req.TestReportURL),
		PackageX86URL:     strings.TrimSpace(req.PackageX86URL),
		PackageARMURL:     strings.TrimSpace(req.PackageARMURL),
		ReleasedAt:        releasedAt,
		VersionConstraint: strings.TrimSpace(req.VersionConstraint),
		Publisher:         strings.TrimSpace(req.Publisher),
		Status:            "published",
		PluginNameZh:      strings.TrimSpace(req.NameZh),
		PluginNameEn:      strings.TrimSpace(req.NameEn),
		DescriptionZh:     strings.TrimSpace(req.DescriptionZh),
		DescriptionEn:     strings.TrimSpace(req.DescriptionEn),
		CapabilityZh:      strings.TrimSpace(req.CapabilityZh),
		CapabilityEn:      strings.TrimSpace(req.CapabilityEn),
		OwnerName:         strings.TrimSpace(req.OwnerName),
	}

	compat := make([]model.MarketPluginCompatibility, 0, len(req.CompatibleItems))
	for _, item := range req.CompatibleItems {
		compat = append(compat, model.MarketPluginCompatibility{
			TargetType:        strings.TrimSpace(item.TargetType),
			ProductCode:       strings.TrimSpace(item.ProductCode),
			ProductName:       strings.TrimSpace(item.ProductName),
			VersionConstraint: strings.TrimSpace(item.VersionConstraint),
		})
	}

	if err := s.repo.UpsertVersion(ctx, version, compat); err != nil {
		return err
	}

	plugin.LatestVersion = version.Version
	return s.repo.UpsertPlugin(ctx, plugin)
}

func (s *Service) OfflineVersion(ctx context.Context, releaseID uint) error {
	return s.repo.OfflineVersion(ctx, releaseID)
}

func (s *Service) DeletePlugin(ctx context.Context, pluginID uint) error {
	return s.repo.DeletePlugin(ctx, pluginID)
}

func (s *Service) ListPublishedPlugins(ctx context.Context, req dto.ListPluginsReq) ([]dto.PublishedPluginItem, int64, error) {
	items, total, err := s.repo.ListPublishedPlugins(ctx, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.PublishedPluginItem, 0, len(items))
	for _, item := range items {
		version, versionErr := s.repo.FindLatestPublishedVersion(ctx, item.ID)
		if versionErr != nil {
			continue
		}
		resp = append(resp, dto.PublishedPluginItem{
			ID:              item.ID,
			PluginID:        item.PluginID,
			Code:            item.Code,
			NameZh:          item.NameZh,
			NameEn:          item.NameEn,
			DescriptionZh:   item.DescriptionZh,
			DescriptionEn:   item.DescriptionEn,
			LatestVersion:   version.Version,
			ReleasedAt:      formatTimePtr(version.ReleasedAt),
			PackageX86URL:   version.PackageX86URL,
			PackageARMURL:   version.PackageARMURL,
			CompatibleItems: mapCompatibility(version.CompatibleItems),
		})
	}
	return resp, total, nil
}

func (s *Service) GetPublishedPluginDetail(ctx context.Context, pluginID uint) (*dto.PublishedPluginDetail, error) {
	plugin, versions, err := s.repo.FindPublishedDetail(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	resp := &dto.PublishedPluginDetail{
		Plugin: dto.PluginDetailItem{
			ID:            plugin.ID,
			PluginID:      plugin.PluginID,
			Code:          plugin.Code,
			NameZh:        plugin.NameZh,
			NameEn:        plugin.NameEn,
			DescriptionZh: plugin.DescriptionZh,
			DescriptionEn: plugin.DescriptionEn,
			CapabilityZh:  plugin.CapabilityZh,
			CapabilityEn:  plugin.CapabilityEn,
			OwnerName:     plugin.OwnerName,
		},
		Versions: make([]dto.PublishedVersion, 0, len(versions)),
	}
	for index, version := range versions {
		item := dto.PublishedVersion{
			ReleaseID:         version.ReleaseID,
			Version:           version.Version,
			Publisher:         version.Publisher,
			VersionConstraint: version.VersionConstraint,
			ChangelogZh:       version.ChangelogZh,
			ChangelogEn:       version.ChangelogEn,
			TestReportURL:     version.TestReportURL,
			PackageX86URL:     version.PackageX86URL,
			PackageARMURL:     version.PackageARMURL,
			ReleasedAt:        formatTimePtr(version.ReleasedAt),
			CompatibleItems:   mapCompatibility(version.CompatibleItems),
		}
		resp.Versions = append(resp.Versions, item)
		if index == 0 {
			resp.Release = item
		}
	}
	return resp, nil
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}

func mapCompatibility(items []model.MarketPluginCompatibility) []dto.CompatibilityItem {
	resp := make([]dto.CompatibilityItem, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.CompatibilityItem{
			TargetType:        item.TargetType,
			ProductCode:       item.ProductCode,
			ProductName:       item.ProductName,
			VersionConstraint: item.VersionConstraint,
		})
	}
	return resp
}
