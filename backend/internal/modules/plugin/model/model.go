package model

import (
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"gorm.io/datatypes"
)

type PluginStatus string

const (
	PluginStatusPlanning PluginStatus = "planning"
	PluginStatusActive   PluginStatus = "active"
	PluginStatusOfflined PluginStatus = "offlined"
)

type PluginReleaseType string

const (
	PluginReleaseTypeInitial     PluginReleaseType = "initial"
	PluginReleaseTypeMaintenance PluginReleaseType = "maintenance"
	PluginReleaseTypeOffline     PluginReleaseType = "offline"
)

type PluginReleaseStatus string

const (
	PluginReleaseStatusDraft            PluginReleaseStatus = "draft"
	PluginReleaseStatusReleasePreparing PluginReleaseStatus = "release_preparing"
	PluginReleaseStatusPendingReview    PluginReleaseStatus = "pending_review"
	PluginReleaseStatusApproved         PluginReleaseStatus = "approved"
	PluginReleaseStatusRejected         PluginReleaseStatus = "rejected"
	PluginReleaseStatusReleased         PluginReleaseStatus = "released"
	PluginReleaseStatusOfflined         PluginReleaseStatus = "offlined"
)

type PluginChecklistItem struct {
	TitleZh string `json:"titleZh"`
	TitleEn string `json:"titleEn"`
	Passed  bool   `json:"passed"`
	NoteZh  string `json:"noteZh,omitempty"`
	NoteEn  string `json:"noteEn,omitempty"`
}

type Plugin struct {
	common.BaseModel
	Code           string       `json:"code" gorm:"type:varchar(64);uniqueIndex;not null;comment:plugin code"`
	RepositoryURL  string       `json:"repositoryUrl" gorm:"type:varchar(255);uniqueIndex;not null;comment:git repository"`
	NameZh         string       `json:"nameZh" gorm:"type:varchar(128);not null;comment:plugin name zh"`
	NameEn         string       `json:"nameEn" gorm:"type:varchar(128);not null;comment:plugin name en"`
	DescriptionZh  string       `json:"descriptionZh" gorm:"type:text;not null;comment:description zh"`
	DescriptionEn  string       `json:"descriptionEn" gorm:"type:text;not null;comment:description en"`
	CapabilityZh   string       `json:"capabilityZh" gorm:"type:text;not null;comment:capabilities zh"`
	CapabilityEn   string       `json:"capabilityEn" gorm:"type:text;not null;comment:capabilities en"`
	Owner          string       `json:"owner" gorm:"type:varchar(128);not null;comment:owner"`
	CreatedBy      uint         `json:"createdBy" gorm:"index;comment:creator user id"`
	CurrentStatus  PluginStatus `json:"currentStatus" gorm:"type:varchar(32);default:planning;index;comment:plugin status"`
	LatestVersion  string       `json:"latestVersion" gorm:"type:varchar(64);comment:latest released version"`
	LastReleasedAt *time.Time   `json:"lastReleasedAt" gorm:"comment:last released at"`
}

func (Plugin) TableName() string {
	return "plugins"
}

type PluginRelease struct {
	common.BaseModel
	PluginID             uint                `json:"pluginId" gorm:"index;not null;comment:plugin id"`
	RequestType          PluginReleaseType   `json:"requestType" gorm:"type:varchar(32);index;not null;comment:request type"`
	Status               PluginReleaseStatus `json:"status" gorm:"type:varchar(32);index;not null;comment:release status"`
	SourceReleaseID      *uint               `json:"sourceReleaseId" gorm:"index;comment:source release id"`
	TargetReleaseID      *uint               `json:"targetReleaseId" gorm:"index;comment:target release id for offline"`
	Version              string              `json:"version" gorm:"type:varchar(64);index;comment:version"`
	VersionConstraint    string              `json:"versionConstraint" gorm:"type:varchar(128);comment:version constraint"`
	Publisher            string              `json:"publisher" gorm:"type:varchar(128);comment:publisher"`
	Checklist            datatypes.JSON      `json:"checklist" gorm:"type:json;comment:test checklist"`
	PerformanceSummaryZh string              `json:"performanceSummaryZh" gorm:"type:text;comment:performance zh"`
	PerformanceSummaryEn string              `json:"performanceSummaryEn" gorm:"type:text;comment:performance en"`
	TestReportURL        string              `json:"testReportUrl" gorm:"type:varchar(255);comment:test report url"`
	PackageX86URL        string              `json:"packageX86Url" gorm:"type:varchar(255);comment:x86 package url"`
	PackageArmURL        string              `json:"packageArmUrl" gorm:"type:varchar(255);comment:arm package url"`
	ChangelogZh          string              `json:"changelogZh" gorm:"type:text;comment:changelog zh"`
	ChangelogEn          string              `json:"changelogEn" gorm:"type:text;comment:changelog en"`
	ReviewComment        string              `json:"reviewComment" gorm:"type:text;comment:review comment"`
	OfflineReasonZh      string              `json:"offlineReasonZh" gorm:"type:text;comment:offline reason zh"`
	OfflineReasonEn      string              `json:"offlineReasonEn" gorm:"type:text;comment:offline reason en"`
	IsOfflined           bool                `json:"isOfflined" gorm:"type:tinyint(1);default:0;comment:is offlined"`
	SubmittedAt          *time.Time          `json:"submittedAt" gorm:"comment:submitted at"`
	ApprovedAt           *time.Time          `json:"approvedAt" gorm:"comment:approved at"`
	ReleasedAt           *time.Time          `json:"releasedAt" gorm:"comment:released at"`
	OfflinedAt           *time.Time          `json:"offlinedAt" gorm:"comment:offlined at"`
	CreatedBy            uint                `json:"createdBy" gorm:"index;comment:creator"`
	ReviewerID           *uint               `json:"reviewerId" gorm:"index;comment:reviewer user id"`
	PublisherID          *uint               `json:"publisherId" gorm:"index;comment:publisher user id"`
	ReviewedBy           *uint               `json:"reviewedBy" gorm:"index;comment:reviewer"`
	Plugin               Plugin              `json:"plugin" gorm:"foreignKey:PluginID"`
}

func (PluginRelease) TableName() string {
	return "plugin_releases"
}

type PluginReleaseEvent struct {
	common.BaseModel
	ReleaseID    uint                `json:"releaseId" gorm:"index;not null;comment:release id"`
	FromStatus   PluginReleaseStatus `json:"fromStatus" gorm:"type:varchar(32);comment:from status"`
	ToStatus     PluginReleaseStatus `json:"toStatus" gorm:"type:varchar(32);comment:to status"`
	Action       string              `json:"action" gorm:"type:varchar(64);comment:transition action"`
	OperatorID   uint                `json:"operatorId" gorm:"index;comment:operator id"`
	Comment      string              `json:"comment" gorm:"type:text;comment:comment"`
	SnapshotJSON datatypes.JSON      `json:"snapshotJson" gorm:"type:json;comment:release snapshot"`
}

func (PluginReleaseEvent) TableName() string {
	return "plugin_release_events"
}
