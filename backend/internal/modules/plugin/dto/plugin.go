package dto

import (
	"github.com/CIPFZ/gowebframe/internal/modules/common"
	pluginModel "github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
)

type PluginChecklistItem = pluginModel.PluginChecklistItem

type SearchPluginReq struct {
	common.PageInfo
	CurrentStatus pluginModel.PluginStatus `json:"currentStatus" form:"currentStatus"`
	CreatedBy     uint                     `json:"createdBy" form:"createdBy"`
}

type GetProjectDetailReq struct {
	ID uint `json:"id" binding:"required"`
}

type SearchPublishedPluginReq struct {
	common.PageInfo
}

type CreatePluginReq struct {
	Code          string `json:"code" binding:"required"`
	RepositoryURL string `json:"repositoryUrl" binding:"required"`
	NameZh        string `json:"nameZh" binding:"required"`
	NameEn        string `json:"nameEn" binding:"required"`
	DescriptionZh string `json:"descriptionZh" binding:"required"`
	DescriptionEn string `json:"descriptionEn" binding:"required"`
	CapabilityZh  string `json:"capabilityZh" binding:"required"`
	CapabilityEn  string `json:"capabilityEn" binding:"required"`
	Owner         string `json:"owner" binding:"required"`
}

type UpdatePluginReq struct {
	ID uint `json:"id" binding:"required"`
	CreatePluginReq
	CurrentStatus pluginModel.PluginStatus `json:"currentStatus"`
}

type SearchReleaseReq struct {
	common.PageInfo
	PluginID    uint                            `json:"pluginId" form:"pluginId"`
	RequestType pluginModel.PluginReleaseType   `json:"requestType" form:"requestType"`
	Status      pluginModel.PluginReleaseStatus `json:"status" form:"status"`
	CreatedBy   uint                            `json:"createdBy" form:"createdBy"`
	ReviewerID  uint                            `json:"reviewerId" form:"reviewerId"`
	PublisherID uint                            `json:"publisherId" form:"publisherId"`
}

type CreateReleaseReq struct {
	PluginID             uint                          `json:"pluginId" binding:"required"`
	RequestType          pluginModel.PluginReleaseType `json:"requestType" binding:"required"`
	SourceReleaseID      *uint                         `json:"sourceReleaseId"`
	TargetReleaseID      *uint                         `json:"targetReleaseId"`
	Version              string                        `json:"version"`
	VersionConstraint    string                        `json:"versionConstraint"`
	Publisher            string                        `json:"publisher"`
	ReviewerID           *uint                         `json:"reviewerId"`
	PublisherID          *uint                         `json:"publisherId"`
	Checklist            []PluginChecklistItem         `json:"checklist"`
	PerformanceSummaryZh string                        `json:"performanceSummaryZh"`
	PerformanceSummaryEn string                        `json:"performanceSummaryEn"`
	TestReportURL        string                        `json:"testReportUrl"`
	PackageX86URL        string                        `json:"packageX86Url"`
	PackageArmURL        string                        `json:"packageArmUrl"`
	ChangelogZh          string                        `json:"changelogZh"`
	ChangelogEn          string                        `json:"changelogEn"`
	OfflineReasonZh      string                        `json:"offlineReasonZh"`
	OfflineReasonEn      string                        `json:"offlineReasonEn"`
}

type UpdateReleaseReq struct {
	ID uint `json:"id" binding:"required"`
	CreateReleaseReq
}

type ReleaseAction string

const (
	ReleaseActionSubmitReview ReleaseAction = "submit_review"
	ReleaseActionApprove      ReleaseAction = "approve"
	ReleaseActionReject       ReleaseAction = "reject"
	ReleaseActionRelease      ReleaseAction = "release"
	ReleaseActionRevise       ReleaseAction = "revise"
)

type ReleaseActionReq struct {
	ID            uint          `json:"id" binding:"required"`
	Action        ReleaseAction `json:"action" binding:"required"`
	ReviewComment string        `json:"reviewComment"`
	ReviewerID    *uint         `json:"reviewerId"`
	PublisherID   *uint         `json:"publisherId"`
}

type AssignReleaseReq struct {
	ID          uint   `json:"id" binding:"required"`
	ReviewerID  *uint  `json:"reviewerId"`
	PublisherID *uint  `json:"publisherId"`
	Comment     string `json:"comment"`
}

type GetReleaseDetailReq struct {
	ID uint `json:"id" binding:"required"`
}

type PluginListItem struct {
	ID                     uint                            `json:"ID"`
	Code                   string                          `json:"code"`
	RepositoryURL          string                          `json:"repositoryUrl"`
	NameZh                 string                          `json:"nameZh"`
	NameEn                 string                          `json:"nameEn"`
	DescriptionZh          string                          `json:"descriptionZh"`
	DescriptionEn          string                          `json:"descriptionEn"`
	CapabilityZh           string                          `json:"capabilityZh"`
	CapabilityEn           string                          `json:"capabilityEn"`
	Owner                  string                          `json:"owner"`
	CreatedBy              uint                            `json:"createdBy"`
	CurrentStatus          pluginModel.PluginStatus        `json:"currentStatus"`
	LatestVersion          string                          `json:"latestVersion"`
	LastReleasedAt         *string                         `json:"lastReleasedAt,omitempty"`
	ReleaseCount           int64                           `json:"releaseCount"`
	PreparingCount         int64                           `json:"preparingCount"`
	PendingReviewCount     int64                           `json:"pendingReviewCount"`
	ApprovedCount          int64                           `json:"approvedCount"`
	PublishedCount         int64                           `json:"publishedCount"`
	OfflinedCount          int64                           `json:"offlinedCount"`
	CurrentWorkflowID      *uint                           `json:"currentWorkflowId,omitempty"`
	CurrentWorkflowType    pluginModel.PluginReleaseType   `json:"currentWorkflowType,omitempty"`
	CurrentWorkflowStatus  pluginModel.PluginReleaseStatus `json:"currentWorkflowStatus,omitempty"`
	CurrentWorkflowVersion string                          `json:"currentWorkflowVersion,omitempty"`
}

type PluginOverview struct {
	ProjectCount       int64 `json:"projectCount"`
	PreparingCount     int64 `json:"preparingCount"`
	PendingReviewCount int64 `json:"pendingReviewCount"`
	ApprovedCount      int64 `json:"approvedCount"`
	PublishedCount     int64 `json:"publishedCount"`
	OfflinedCount      int64 `json:"offlinedCount"`
}

type ReleaseListItem struct {
	ID                   uint                            `json:"ID"`
	PluginID             uint                            `json:"pluginId"`
	PluginCode           string                          `json:"pluginCode"`
	PluginNameZh         string                          `json:"pluginNameZh"`
	PluginNameEn         string                          `json:"pluginNameEn"`
	RequestType          pluginModel.PluginReleaseType   `json:"requestType"`
	Status               pluginModel.PluginReleaseStatus `json:"status"`
	Version              string                          `json:"version"`
	VersionConstraint    string                          `json:"versionConstraint"`
	Publisher            string                          `json:"publisher"`
	ReviewerID           *uint                           `json:"reviewerId"`
	PublisherID          *uint                           `json:"publisherId"`
	Checklist            []PluginChecklistItem           `json:"checklist"`
	PerformanceSummaryZh string                          `json:"performanceSummaryZh"`
	PerformanceSummaryEn string                          `json:"performanceSummaryEn"`
	TestReportURL        string                          `json:"testReportUrl"`
	PackageX86URL        string                          `json:"packageX86Url"`
	PackageArmURL        string                          `json:"packageArmUrl"`
	ChangelogZh          string                          `json:"changelogZh"`
	ChangelogEn          string                          `json:"changelogEn"`
	ReviewComment        string                          `json:"reviewComment"`
	OfflineReasonZh      string                          `json:"offlineReasonZh"`
	OfflineReasonEn      string                          `json:"offlineReasonEn"`
	IsOfflined           bool                            `json:"isOfflined"`
	SourceReleaseID      *uint                           `json:"sourceReleaseId"`
	TargetReleaseID      *uint                           `json:"targetReleaseId"`
	CreatedBy            uint                            `json:"createdBy"`
	CreatedAt            string                          `json:"createdAt"`
	SubmittedAt          *string                         `json:"submittedAt,omitempty"`
	ApprovedAt           *string                         `json:"approvedAt,omitempty"`
	ReleasedAt           *string                         `json:"releasedAt,omitempty"`
	OfflinedAt           *string                         `json:"offlinedAt,omitempty"`
}

type ReleaseEventItem struct {
	ID         uint   `json:"ID"`
	ReleaseID  uint   `json:"releaseId"`
	FromStatus string `json:"fromStatus"`
	ToStatus   string `json:"toStatus"`
	Action     string `json:"action"`
	OperatorID uint   `json:"operatorId"`
	Comment    string `json:"comment"`
	CreatedAt  string `json:"createdAt"`
}

type ReleaseDetail struct {
	ReleaseListItem
	Events []ReleaseEventItem `json:"events"`
}

type ProjectDetail struct {
	PluginListItem
	CurrentWorkflow    *ReleaseListItem  `json:"currentWorkflow,omitempty"`
	LatestReleased     *ReleaseListItem  `json:"latestReleased,omitempty"`
	PreparingCount     int64             `json:"preparingCount"`
	PendingReviewCount int64             `json:"pendingReviewCount"`
	ApprovedCount      int64             `json:"approvedCount"`
	PublishedCount     int64             `json:"publishedCount"`
	OfflinedCount      int64             `json:"offlinedCount"`
	Versions           []ReleaseListItem `json:"versions"`
}

type PublishedPluginListItem struct {
	PluginID             uint   `json:"pluginId"`
	ReleaseID            uint   `json:"releaseId"`
	Code                 string `json:"code"`
	NameZh               string `json:"nameZh"`
	NameEn               string `json:"nameEn"`
	DescriptionZh        string `json:"descriptionZh"`
	DescriptionEn        string `json:"descriptionEn"`
	CapabilityZh         string `json:"capabilityZh"`
	CapabilityEn         string `json:"capabilityEn"`
	Owner                string `json:"owner"`
	Version              string `json:"version"`
	VersionConstraint    string `json:"versionConstraint"`
	Publisher            string `json:"publisher"`
	PackageX86URL        string `json:"packageX86Url"`
	PackageArmURL        string `json:"packageArmUrl"`
	TestReportURL        string `json:"testReportUrl"`
	ChangelogZh          string `json:"changelogZh"`
	ChangelogEn          string `json:"changelogEn"`
	PerformanceSummaryZh string `json:"performanceSummaryZh"`
	PerformanceSummaryEn string `json:"performanceSummaryEn"`
	ReleasedAt           string `json:"releasedAt"`
}

type GetPublishedPluginDetailReq struct {
	PluginID uint `json:"pluginId" binding:"required"`
}

type PublishedPluginVersionItem struct {
	ReleaseID            uint   `json:"releaseId"`
	Version              string `json:"version"`
	VersionConstraint    string `json:"versionConstraint"`
	Publisher            string `json:"publisher"`
	PackageX86URL        string `json:"packageX86Url"`
	PackageArmURL        string `json:"packageArmUrl"`
	TestReportURL        string `json:"testReportUrl"`
	ChangelogZh          string `json:"changelogZh"`
	ChangelogEn          string `json:"changelogEn"`
	PerformanceSummaryZh string `json:"performanceSummaryZh"`
	PerformanceSummaryEn string `json:"performanceSummaryEn"`
	ReleasedAt           string `json:"releasedAt"`
}

type PublishedPluginDetail struct {
	PluginID      uint                         `json:"pluginId"`
	Code          string                       `json:"code"`
	NameZh        string                       `json:"nameZh"`
	NameEn        string                       `json:"nameEn"`
	DescriptionZh string                       `json:"descriptionZh"`
	DescriptionEn string                       `json:"descriptionEn"`
	CapabilityZh  string                       `json:"capabilityZh"`
	CapabilityEn  string                       `json:"capabilityEn"`
	Owner         string                       `json:"owner"`
	Versions      []PublishedPluginVersionItem `json:"versions"`
}
