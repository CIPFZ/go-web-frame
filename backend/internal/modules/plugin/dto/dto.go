package dto

import "github.com/CIPFZ/gowebframe/internal/modules/common"

type PageInfo = common.PageInfo

const (
	WorkOrderScopeAll  = "all"
	WorkOrderScopeMine = "mine"
)

type CreatePluginReq struct {
	Code          string `json:"code" binding:"required"`
	RepositoryURL string `json:"repositoryUrl" binding:"required"`
	NameZh        string `json:"nameZh" binding:"required"`
	NameEn        string `json:"nameEn" binding:"required"`
	DescriptionZh string `json:"descriptionZh" binding:"required"`
	DescriptionEn string `json:"descriptionEn" binding:"required"`
	DepartmentID  uint   `json:"departmentId" binding:"required"`
	OwnerID       uint   `json:"ownerId"`
}

type UpdatePluginReq struct {
	ID            uint   `json:"id" binding:"required"`
	RepositoryURL string `json:"repositoryUrl" binding:"required"`
	NameZh        string `json:"nameZh" binding:"required"`
	NameEn        string `json:"nameEn" binding:"required"`
	DescriptionZh string `json:"descriptionZh" binding:"required"`
	DescriptionEn string `json:"descriptionEn" binding:"required"`
	DepartmentID  uint   `json:"departmentId" binding:"required"`
	OwnerID       uint   `json:"ownerId"`
}

type SearchPluginReq struct {
	Code string `json:"code"`
	Name string `json:"name"`
	PageInfo
}

type UpsertCompatibleProductReq struct {
	ProductID         uint   `json:"productId" binding:"required"`
	VersionConstraint string `json:"versionConstraint"`
}

type ReleaseCompatibilityReq struct {
	ProductItems []UpsertCompatibleProductReq `json:"productItems"`
	AcliItems    []UpsertCompatibleProductReq `json:"acliItems"`
	Universal    bool                         `json:"universal"`
}

type CreateReleaseReq struct {
	PluginID        uint                    `json:"pluginId" binding:"required"`
	RequestType     int8                    `json:"requestType" binding:"required"`
	Version         string                  `json:"version"`
	TestReportURL   string                  `json:"testReportUrl"`
	PackageX86URL   string                  `json:"packageX86Url"`
	PackageARMURL   string                  `json:"packageArmUrl"`
	ChangelogZh     string                  `json:"changelogZh"`
	ChangelogEn     string                  `json:"changelogEn"`
	OfflineReasonZh string                  `json:"offlineReasonZh"`
	OfflineReasonEn string                  `json:"offlineReasonEn"`
	TDID            string                  `json:"tdId"`
	Compatibility   ReleaseCompatibilityReq `json:"compatibility"`
}

type UpdateReleaseReq struct {
	ID              uint                    `json:"id" binding:"required"`
	Version         string                  `json:"version"`
	TestReportURL   string                  `json:"testReportUrl"`
	PackageX86URL   string                  `json:"packageX86Url"`
	PackageARMURL   string                  `json:"packageArmUrl"`
	ChangelogZh     string                  `json:"changelogZh"`
	ChangelogEn     string                  `json:"changelogEn"`
	OfflineReasonZh string                  `json:"offlineReasonZh"`
	OfflineReasonEn string                  `json:"offlineReasonEn"`
	TDID            string                  `json:"tdId"`
	Compatibility   ReleaseCompatibilityReq `json:"compatibility"`
}

type TransitionReleaseReq struct {
	ID              uint   `json:"id" binding:"required"`
	Action          string `json:"action" binding:"required"`
	ReviewComment   string `json:"reviewComment"`
	OfflineReasonZh string `json:"offlineReasonZh"`
	OfflineReasonEn string `json:"offlineReasonEn"`
	TDID            string `json:"tdId"`
}

type ClaimWorkOrderReq struct {
	ID uint `json:"id" binding:"required"`
}

type ResetWorkOrderReq struct {
	ID     uint   `json:"id" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

type SearchWorkOrderReq struct {
	Scope         string `json:"scope"`
	Keyword       string `json:"keyword"`
	ProcessStatus *int8  `json:"processStatus"`
	Status        *int8  `json:"status"`
	RequestType   *int8  `json:"requestType"`
	ClaimerID     *uint  `json:"claimerId"`
	PluginID      *uint  `json:"pluginId"`
	PageInfo
}

type GetProjectDetailReq struct {
	ID        uint  `json:"id" binding:"required"`
	ReleaseID *uint `json:"releaseId"`
}

type GetReleaseDetailReq struct {
	ID uint `json:"id" binding:"required"`
}

type SearchProductReq struct {
	IncludeInactive bool `json:"includeInactive"`
	PageInfo
}

type CreateProductReq struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
}

type UpdateProductReq struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
	Status      bool   `json:"status"`
}

type SearchDepartmentReq struct {
	IncludeInactive bool `json:"includeInactive"`
	PageInfo
}

type CreateDepartmentReq struct {
	NameZh      string `json:"nameZh" binding:"required"`
	NameEn      string `json:"nameEn" binding:"required"`
	ProductLine string `json:"productLine" binding:"required"`
	ParentID    *uint  `json:"parentId"`
}

type UpdateDepartmentReq struct {
	ID          uint   `json:"id" binding:"required"`
	NameZh      string `json:"nameZh" binding:"required"`
	NameEn      string `json:"nameEn" binding:"required"`
	ProductLine string `json:"productLine" binding:"required"`
	ParentID    *uint  `json:"parentId"`
	Status      bool   `json:"status"`
}

type GetPublishedPluginListReq struct {
	PageInfo
}

type DepartmentItem struct {
	ID          uint   `json:"ID"`
	Name        string `json:"name"`
	NameZh      string `json:"nameZh"`
	NameEn      string `json:"nameEn"`
	ProductLine string `json:"productLine"`
	Status      bool   `json:"status"`
}

type ProductItem struct {
	ID          uint   `json:"ID"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Status      bool   `json:"status"`
}

type CompatibleProductItem struct {
	ID                uint   `json:"ID"`
	ProductID         uint   `json:"productId"`
	ProductCode       string `json:"productCode"`
	ProductName       string `json:"productName"`
	Type              string `json:"type"`
	VersionConstraint string `json:"versionConstraint"`
}

type ReleaseCompatibility struct {
	ProductItems []CompatibleProductItem `json:"productItems"`
	AcliItems    []CompatibleProductItem `json:"acliItems"`
	Universal    bool                    `json:"universal"`
}

type EventItem struct {
	ID                uint   `json:"ID"`
	Action            string `json:"action"`
	FromStatus        int8   `json:"fromStatus"`
	ToStatus          int8   `json:"toStatus"`
	FromProcessStatus int8   `json:"fromProcessStatus"`
	ToProcessStatus   int8   `json:"toProcessStatus"`
	OperatorID        uint   `json:"operatorId"`
	Comment           string `json:"comment"`
	CreatedAt         string `json:"createdAt"`
}

type PluginItem struct {
	ID            uint   `json:"ID"`
	Code          string `json:"code"`
	RepositoryURL string `json:"repositoryUrl"`
	NameZh        string `json:"nameZh"`
	NameEn        string `json:"nameEn"`
	DescriptionZh string `json:"descriptionZh"`
	DescriptionEn string `json:"descriptionEn"`
	DepartmentID  uint   `json:"departmentId"`
	Department    string `json:"department"`
	OwnerID       uint   `json:"ownerId"`
	CreatedBy     uint   `json:"createdBy"`
	CreatedAt     string `json:"createdAt"`
}

type PluginReleaseItem struct {
	ID              uint                    `json:"ID"`
	PluginID        uint                    `json:"pluginId"`
	PluginCode      string                  `json:"pluginCode"`
	PluginNameZh    string                  `json:"pluginNameZh"`
	RequestType     int8                    `json:"requestType"`
	Status          int8                    `json:"status"`
	ProcessStatus   int8                    `json:"processStatus"`
	Version         string                  `json:"version"`
	ClaimerID       *uint                   `json:"claimerId"`
	ClaimerName     string                  `json:"claimerName"`
	ClaimerUsername string                  `json:"claimerUsername"`
	ReviewComment   string                  `json:"reviewComment"`
	TestReportURL   string                  `json:"testReportUrl"`
	PackageX86URL   string                  `json:"packageX86Url"`
	PackageARMURL   string                  `json:"packageArmUrl"`
	ChangelogZh     string                  `json:"changelogZh"`
	ChangelogEn     string                  `json:"changelogEn"`
	OfflineReasonZh string                  `json:"offlineReasonZh"`
	OfflineReasonEn string                  `json:"offlineReasonEn"`
	TDID            string                  `json:"tdId"`
	SubmittedAt     *string                 `json:"submittedAt"`
	ApprovedAt      *string                 `json:"approvedAt"`
	ReleasedAt      *string                 `json:"releasedAt"`
	OfflinedAt      *string                 `json:"offlinedAt"`
	ClaimedAt       *string                 `json:"claimedAt"`
	Compatibility   ReleaseCompatibility    `json:"compatibility"`
	CompatibleItems []CompatibleProductItem `json:"compatibleItems"`
	CreatedBy       uint                    `json:"createdBy"`
	CreatedAt       string                  `json:"createdAt"`
}

type WorkOrderItem struct {
	PluginReleaseItem
}

type ProjectDetail struct {
	Plugin          PluginItem          `json:"plugin"`
	SelectedRelease *PluginReleaseItem  `json:"selectedRelease,omitempty"`
	Releases        []PluginReleaseItem `json:"releases"`
	Events          []EventItem         `json:"events"`
}

type PublishedPluginItem struct {
	ID              uint                    `json:"ID"`
	Code            string                  `json:"code"`
	NameZh          string                  `json:"nameZh"`
	NameEn          string                  `json:"nameEn"`
	DescriptionZh   string                  `json:"descriptionZh"`
	DescriptionEn   string                  `json:"descriptionEn"`
	LatestVersion   string                  `json:"latestVersion"`
	CompatibleItems []CompatibleProductItem `json:"compatibleItems"`
}

type PublishedPluginDetail struct {
	Plugin   PluginItem             `json:"plugin"`
	Release  PublishedReleaseItem   `json:"release"`
	Versions []PublishedReleaseItem `json:"versions"`
}

type PublishedReleaseItem struct {
	ID              uint                    `json:"ID"`
	Version         string                  `json:"version"`
	ChangelogZh     string                  `json:"changelogZh"`
	ChangelogEn     string                  `json:"changelogEn"`
	TestReportURL   string                  `json:"testReportUrl"`
	PackageX86URL   string                  `json:"packageX86Url"`
	PackageARMURL   string                  `json:"packageArmUrl"`
	ReleasedAt      *string                 `json:"releasedAt"`
	CompatibleItems []CompatibleProductItem `json:"compatibleItems"`
}
