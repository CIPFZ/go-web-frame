package dto

type PageInfo struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type CompatibilityItem struct {
	TargetType        string `json:"targetType"`
	ProductCode       string `json:"productCode"`
	ProductName       string `json:"productName"`
	VersionConstraint string `json:"versionConstraint"`
}

type UpsertPluginReq struct {
	PluginID      uint   `json:"pluginId" binding:"required"`
	Code          string `json:"code" binding:"required"`
	NameZh        string `json:"nameZh" binding:"required"`
	NameEn        string `json:"nameEn" binding:"required"`
	DescriptionZh string `json:"descriptionZh" binding:"required"`
	DescriptionEn string `json:"descriptionEn" binding:"required"`
	CapabilityZh  string `json:"capabilityZh"`
	CapabilityEn  string `json:"capabilityEn"`
	OwnerName     string `json:"ownerName"`
	Status        string `json:"status"`
}

type UpsertVersionReq struct {
	PluginID          uint                `json:"pluginId" binding:"required"`
	ReleaseID         uint                `json:"releaseId" binding:"required"`
	Version           string              `json:"version" binding:"required"`
	NameZh            string              `json:"nameZh" binding:"required"`
	NameEn            string              `json:"nameEn" binding:"required"`
	DescriptionZh     string              `json:"descriptionZh" binding:"required"`
	DescriptionEn     string              `json:"descriptionEn" binding:"required"`
	CapabilityZh      string              `json:"capabilityZh"`
	CapabilityEn      string              `json:"capabilityEn"`
	OwnerName         string              `json:"ownerName"`
	Publisher         string              `json:"publisher"`
	ChangelogZh       string              `json:"changelogZh"`
	ChangelogEn       string              `json:"changelogEn"`
	TestReportURL     string              `json:"testReportUrl"`
	PackageX86URL     string              `json:"packageX86Url"`
	PackageARMURL     string              `json:"packageArmUrl"`
	ReleasedAt        string              `json:"releasedAt"`
	VersionConstraint string              `json:"versionConstraint"`
	CompatibleItems   []CompatibilityItem `json:"compatibleItems"`
}

type OfflineVersionReq struct {
	ReleaseID uint `json:"releaseId" binding:"required"`
}

type DeletePluginReq struct {
	PluginID uint `json:"pluginId" binding:"required"`
}

type ListPluginsReq struct {
	Keyword string `json:"keyword"`
	PageInfo
}

type GetPluginDetailReq struct {
	PluginID uint `json:"pluginId" binding:"required"`
}

type PublishedPluginItem struct {
	ID              uint                `json:"ID"`
	PluginID        uint                `json:"pluginId"`
	Code            string              `json:"code"`
	NameZh          string              `json:"nameZh"`
	NameEn          string              `json:"nameEn"`
	DescriptionZh   string              `json:"descriptionZh"`
	DescriptionEn   string              `json:"descriptionEn"`
	LatestVersion   string              `json:"latestVersion"`
	ReleasedAt      *string             `json:"releasedAt"`
	PackageX86URL   string              `json:"packageX86Url"`
	PackageARMURL   string              `json:"packageArmUrl"`
	CompatibleItems []CompatibilityItem `json:"compatibleItems"`
}

type PublishedPluginDetail struct {
	Plugin PluginDetailItem     `json:"plugin"`
	Release PublishedVersion    `json:"release"`
	Versions []PublishedVersion `json:"versions"`
}

type PluginDetailItem struct {
	ID            uint   `json:"ID"`
	PluginID      uint   `json:"pluginId"`
	Code          string `json:"code"`
	NameZh        string `json:"nameZh"`
	NameEn        string `json:"nameEn"`
	DescriptionZh string `json:"descriptionZh"`
	DescriptionEn string `json:"descriptionEn"`
	CapabilityZh  string `json:"capabilityZh"`
	CapabilityEn  string `json:"capabilityEn"`
	OwnerName     string `json:"ownerName"`
}

type PublishedVersion struct {
	ReleaseID         uint                `json:"releaseId"`
	Version           string              `json:"version"`
	Publisher         string              `json:"publisher"`
	VersionConstraint string              `json:"versionConstraint"`
	ChangelogZh       string              `json:"changelogZh"`
	ChangelogEn       string              `json:"changelogEn"`
	TestReportURL     string              `json:"testReportUrl"`
	PackageX86URL     string              `json:"packageX86Url"`
	PackageARMURL     string              `json:"packageArmUrl"`
	ReleasedAt        *string             `json:"releasedAt"`
	CompatibleItems   []CompatibilityItem `json:"compatibleItems"`
}
