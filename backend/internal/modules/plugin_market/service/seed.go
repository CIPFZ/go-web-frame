package service

import (
	"context"

	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/dto"
)

type demoPluginSeed struct {
	plugin   dto.UpsertPluginReq
	versions []dto.UpsertVersionReq
}

func (s *Service) SeedDemoDataIfEmpty(ctx context.Context) error {
	total, err := s.repo.CountPlugins(ctx)
	if err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	for _, item := range defaultDemoSeeds() {
		if err := s.UpsertPlugin(ctx, item.plugin); err != nil {
			return err
		}
		for _, version := range item.versions {
			if err := s.UpsertVersion(ctx, version); err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultDemoSeeds() []demoPluginSeed {
	return []demoPluginSeed{
		{
			plugin: dto.UpsertPluginReq{
				PluginID:      1001,
				Code:          "agent-helper",
				NameZh:        "智能助手插件",
				NameEn:        "Agent Helper",
				DescriptionZh: "面向故障排查与自动巡检场景的智能辅助插件，提供任务编排、日志诊断与常见问题建议。",
				DescriptionEn: "An assistant plugin for troubleshooting and automated inspections with task orchestration, log diagnosis, and guided suggestions.",
				CapabilityZh:  "支持自动巡检、日志摘要、风险提示与诊断建议输出。",
				CapabilityEn:  "Provides automated inspection, log summarization, risk alerts, and diagnosis suggestions.",
				OwnerName:     "平台插件组",
			},
			versions: []dto.UpsertVersionReq{
				{
					PluginID:          1001,
					ReleaseID:         50011,
					Version:           "1.0.0",
					NameZh:            "智能助手插件",
					NameEn:            "Agent Helper",
					DescriptionZh:     "面向故障排查与自动巡检场景的智能辅助插件，提供任务编排、日志诊断与常见问题建议。",
					DescriptionEn:     "An assistant plugin for troubleshooting and automated inspections with task orchestration, log diagnosis, and guided suggestions.",
					CapabilityZh:      "支持自动巡检、日志摘要、风险提示与诊断建议输出。",
					CapabilityEn:      "Provides automated inspection, log summarization, risk alerts, and diagnosis suggestions.",
					OwnerName:         "平台插件组",
					Publisher:         "Plugin Market Demo",
					ChangelogZh:       "首个公开演示版本，提供基础巡检和日志诊断能力。",
					ChangelogEn:       "Initial public demo release with baseline inspection and log diagnosis capabilities.",
					TestReportURL:     "https://downloads.example.com/plugin-market/agent-helper/reports/1.0.0.pdf",
					PackageX86URL:     "https://downloads.example.com/plugin-market/agent-helper/1.0.0/agent-helper-x86_64.zip",
					PackageARMURL:     "https://downloads.example.com/plugin-market/agent-helper/1.0.0/agent-helper-arm64.zip",
					ReleasedAt:        "2026-04-10T09:00:00Z",
					VersionConstraint: ">=1.0.0",
					CompatibleItems: []dto.CompatibilityItem{
						{TargetType: "product", ProductCode: "DTP", ProductName: "故障排查平台", VersionConstraint: ">=2.0.0"},
						{TargetType: "acli", ProductCode: "ACLI", ProductName: "aCLI", VersionConstraint: ">=2.3.0"},
					},
				},
				{
					PluginID:          1001,
					ReleaseID:         50012,
					Version:           "1.1.0",
					NameZh:            "智能助手插件",
					NameEn:            "Agent Helper",
					DescriptionZh:     "面向故障排查与自动巡检场景的智能辅助插件，提供任务编排、日志诊断与常见问题建议。",
					DescriptionEn:     "An assistant plugin for troubleshooting and automated inspections with task orchestration, log diagnosis, and guided suggestions.",
					CapabilityZh:      "支持自动巡检、日志摘要、风险提示与诊断建议输出。",
					CapabilityEn:      "Provides automated inspection, log summarization, risk alerts, and diagnosis suggestions.",
					OwnerName:         "平台插件组",
					Publisher:         "Plugin Market Demo",
					ChangelogZh:       "增强多阶段巡检流程，补充 ARM64 安装包与测试报告。",
					ChangelogEn:       "Enhanced multi-stage inspection flows and added ARM64 packages plus test reports.",
					TestReportURL:     "https://downloads.example.com/plugin-market/agent-helper/reports/1.1.0.pdf",
					PackageX86URL:     "https://downloads.example.com/plugin-market/agent-helper/1.1.0/agent-helper-x86_64.zip",
					PackageARMURL:     "https://downloads.example.com/plugin-market/agent-helper/1.1.0/agent-helper-arm64.zip",
					ReleasedAt:        "2026-04-16T09:00:00Z",
					VersionConstraint: ">=1.0.0",
					CompatibleItems: []dto.CompatibilityItem{
						{TargetType: "product", ProductCode: "DTP", ProductName: "故障排查平台", VersionConstraint: ">=2.1.0"},
						{TargetType: "acli", ProductCode: "ACLI", ProductName: "aCLI", VersionConstraint: ">=2.4.0"},
					},
				},
			},
		},
		{
			plugin: dto.UpsertPluginReq{
				PluginID:      1002,
				Code:          "image-optimizer",
				NameZh:        "图像优化插件",
				NameEn:        "Image Optimizer",
				DescriptionZh: "用于镜像压缩、格式转换与制品瘦身的构建辅助插件，适合发布前质量处理。",
				DescriptionEn: "A build helper plugin for image compression, format conversion, and artifact slimming before releases.",
				CapabilityZh:  "支持镜像压缩、格式标准化和批量优化处理。",
				CapabilityEn:  "Supports image compression, format normalization, and batch optimization pipelines.",
				OwnerName:     "视觉工程组",
			},
			versions: []dto.UpsertVersionReq{
				{
					PluginID:          1002,
					ReleaseID:         50021,
					Version:           "2.0.0",
					NameZh:            "图像优化插件",
					NameEn:            "Image Optimizer",
					DescriptionZh:     "用于镜像压缩、格式转换与制品瘦身的构建辅助插件，适合发布前质量处理。",
					DescriptionEn:     "A build helper plugin for image compression, format conversion, and artifact slimming before releases.",
					CapabilityZh:      "支持镜像压缩、格式标准化和批量优化处理。",
					CapabilityEn:      "Supports image compression, format normalization, and batch optimization pipelines.",
					OwnerName:         "视觉工程组",
					Publisher:         "Plugin Market Demo",
					ChangelogZh:       "新增批量优化任务编排，支持产品安装包前置压缩。",
					ChangelogEn:       "Adds batch optimization workflows for pre-release packaging.",
					TestReportURL:     "https://downloads.example.com/plugin-market/image-optimizer/reports/2.0.0.pdf",
					PackageX86URL:     "https://downloads.example.com/plugin-market/image-optimizer/2.0.0/image-optimizer-x86_64.zip",
					ReleasedAt:        "2026-04-12T08:30:00Z",
					VersionConstraint: ">=2.0.0",
					CompatibleItems: []dto.CompatibilityItem{
						{TargetType: "product", ProductCode: "CMS", ProductName: "插件管理中心", VersionConstraint: ">=2.0.0"},
					},
				},
			},
		},
		{
			plugin: dto.UpsertPluginReq{
				PluginID:      1003,
				Code:          "edge-runtime",
				NameZh:        "边缘运行时插件",
				NameEn:        "Edge Runtime",
				DescriptionZh: "为边缘节点提供轻量执行环境与状态回传能力，适用于多架构部署场景。",
				DescriptionEn: "Provides a lightweight execution runtime and status reporting for edge nodes in multi-architecture deployments.",
				CapabilityZh:  "支持边缘节点注册、任务下发和运行状态同步。",
				CapabilityEn:  "Supports node registration, task delivery, and runtime status synchronization.",
				OwnerName:     "边缘计算组",
			},
			versions: []dto.UpsertVersionReq{
				{
					PluginID:          1003,
					ReleaseID:         50031,
					Version:           "1.5.2",
					NameZh:            "边缘运行时插件",
					NameEn:            "Edge Runtime",
					DescriptionZh:     "为边缘节点提供轻量执行环境与状态回传能力，适用于多架构部署场景。",
					DescriptionEn:     "Provides a lightweight execution runtime and status reporting for edge nodes in multi-architecture deployments.",
					CapabilityZh:      "支持边缘节点注册、任务下发和运行状态同步。",
					CapabilityEn:      "Supports node registration, task delivery, and runtime status synchronization.",
					OwnerName:         "边缘计算组",
					Publisher:         "Plugin Market Demo",
					ChangelogZh:       "补充 ARM64 版本，优化边缘状态心跳可靠性。",
					ChangelogEn:       "Adds ARM64 support and improves heartbeat reliability for edge nodes.",
					TestReportURL:     "https://downloads.example.com/plugin-market/edge-runtime/reports/1.5.2.pdf",
					PackageX86URL:     "https://downloads.example.com/plugin-market/edge-runtime/1.5.2/edge-runtime-x86_64.zip",
					PackageARMURL:     "https://downloads.example.com/plugin-market/edge-runtime/1.5.2/edge-runtime-arm64.zip",
					ReleasedAt:        "2026-04-08T13:00:00Z",
					VersionConstraint: ">=1.4.0",
					CompatibleItems: []dto.CompatibilityItem{
						{TargetType: "product", ProductCode: "EDGE", ProductName: "边缘节点网关", VersionConstraint: ">=3.2.0"},
						{TargetType: "acli", ProductCode: "ACLI", ProductName: "aCLI", VersionConstraint: ">=2.2.0"},
					},
				},
			},
		},
		{
			plugin: dto.UpsertPluginReq{
				PluginID:      1004,
				Code:          "report-studio",
				NameZh:        "报表分析插件",
				NameEn:        "Report Studio",
				DescriptionZh: "面向插件运营与质量追踪的报表中心，提供版本、下载与兼容性数据统计。",
				DescriptionEn: "A reporting hub for plugin operations and quality tracking with release, download, and compatibility analytics.",
				CapabilityZh:  "支持版本统计、下载趋势分析与兼容矩阵展示。",
				CapabilityEn:  "Supports version analytics, download trend tracking, and compatibility matrix views.",
				OwnerName:     "数据分析组",
			},
			versions: []dto.UpsertVersionReq{
				{
					PluginID:          1004,
					ReleaseID:         50041,
					Version:           "3.2.1",
					NameZh:            "报表分析插件",
					NameEn:            "Report Studio",
					DescriptionZh:     "面向插件运营与质量追踪的报表中心，提供版本、下载与兼容性数据统计。",
					DescriptionEn:     "A reporting hub for plugin operations and quality tracking with release, download, and compatibility analytics.",
					CapabilityZh:      "支持版本统计、下载趋势分析与兼容矩阵展示。",
					CapabilityEn:      "Supports version analytics, download trend tracking, and compatibility matrix views.",
					OwnerName:         "数据分析组",
					Publisher:         "Plugin Market Demo",
					ChangelogZh:       "新增插件下载趋势面板和兼容性矩阵导出。",
					ChangelogEn:       "Introduces download trend dashboards and compatibility matrix export.",
					TestReportURL:     "https://downloads.example.com/plugin-market/report-studio/reports/3.2.1.pdf",
					PackageX86URL:     "https://downloads.example.com/plugin-market/report-studio/3.2.1/report-studio-x86_64.zip",
					ReleasedAt:        "2026-04-14T11:20:00Z",
					VersionConstraint: ">=3.0.0",
					CompatibleItems: []dto.CompatibilityItem{
						{TargetType: "product", ProductCode: "PMC", ProductName: "插件管理中心", VersionConstraint: ">=2.1.0"},
					},
				},
			},
		},
	}
}
