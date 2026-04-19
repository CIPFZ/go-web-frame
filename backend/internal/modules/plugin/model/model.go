package model

import (
	"time"

	"github.com/CIPFZ/gowebframe/internal/modules/common"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

const (
	ReleaseRequestTypeVersion int8 = 1
	ReleaseRequestTypeOffline int8 = 2
)

const (
	ReleaseStatusDraft         int8 = 0
	ReleaseStatusReady         int8 = 1
	ReleaseStatusPendingReview int8 = 2
	ReleaseStatusApproved      int8 = 3
	ReleaseStatusRejected      int8 = 4
	ReleaseStatusReleased      int8 = 5
	ReleaseStatusOfflined      int8 = 6
)

const (
	ReleaseProcessStatusPending    int8 = 0
	ReleaseProcessStatusProcessing int8 = 1
	ReleaseProcessStatusRejected   int8 = 2
	ReleaseProcessStatusDone       int8 = 3
)

const (
	ReleaseActionCreate         = "create"
	ReleaseActionSubmitReview   = "submit_review"
	ReleaseActionApprove        = "approve"
	ReleaseActionReject         = "reject"
	ReleaseActionRelease        = "release"
	ReleaseActionRevise         = "revise"
	ReleaseActionRequestOffline = "request_offline"
	ReleaseActionOffline        = "offline"
	ReleaseActionClaim          = "claim"
	ReleaseActionReset          = "reset"
)

const (
	CompatibleTargetTypeProduct = "product"
	CompatibleTargetTypeAcli    = "acli"
)

const (
	CompatibilityTypeProduct   = CompatibleTargetTypeProduct
	CompatibilityTypeAcli      = CompatibleTargetTypeAcli
	CompatibilityTypeUniversal = "universal"
)

type PluginDepartment struct {
	common.BaseModel
	Name        string `json:"name" gorm:"type:varchar(64);not null"`
	NameZh      string `json:"nameZh" gorm:"type:varchar(64);default:''"`
	NameEn      string `json:"nameEn" gorm:"type:varchar(64);default:''"`
	ProductLine string `json:"productLine" gorm:"type:varchar(64);not null"`
	ParentID    *uint  `json:"parentId"`
	Sort        int    `json:"sort" gorm:"default:0"`
	Status      bool   `json:"status" gorm:"default:true"`
}

func (PluginDepartment) TableName() string { return "sys_departments" }

type PluginProduct struct {
	common.BaseModel
	Code        string `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"`
	Name        string `json:"name" gorm:"type:varchar(128);not null"`
	Type        string `json:"type" gorm:"type:varchar(16);default:'product';index;not null"`
	Description string `json:"description" gorm:"type:text"`
	Sort        int    `json:"sort" gorm:"default:0"`
	Status      bool   `json:"status" gorm:"default:true"`
}

func (PluginProduct) TableName() string { return "sys_products" }

type Plugin struct {
	common.BaseModel
	Code          string `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"`
	RepositoryURL string `json:"repositoryUrl" gorm:"type:varchar(255);uniqueIndex;not null"`
	NameZh        string `json:"nameZh" gorm:"type:varchar(128);not null"`
	NameEn        string `json:"nameEn" gorm:"type:varchar(128);not null"`
	DescriptionZh string `json:"descriptionZh" gorm:"type:text;not null"`
	DescriptionEn string `json:"descriptionEn" gorm:"type:text;not null"`
	DepartmentID  uint   `json:"departmentId"`
	OwnerID       uint   `json:"ownerId"`
	CreatedBy     uint   `json:"createdBy"`

	Department PluginDepartment `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	Releases   []PluginRelease  `json:"releases,omitempty" gorm:"foreignKey:PluginID"`
}

func (Plugin) TableName() string { return "plugins" }

type PluginRelease struct {
	common.BaseModel
	PluginID        uint                      `json:"pluginId" gorm:"index;not null"`
	RequestType     int8                      `json:"requestType" gorm:"default:1;not null"`
	Status          int8                      `json:"status" gorm:"default:0;not null"`
	ProcessStatus   int8                      `json:"processStatus" gorm:"default:0;not null"`
	Version         string                    `json:"version" gorm:"type:varchar(64)"`
	Universal       bool                      `json:"universal" gorm:"default:false;not null"`
	ClaimerID       *uint                     `json:"claimerId" gorm:"index"`
	ClaimedAt       *time.Time                `json:"claimedAt"`
	Checklist       string                    `json:"checklist" gorm:"type:text"`
	TestReportURL   string                    `json:"testReportUrl" gorm:"type:varchar(255)"`
	PackageX86URL   string                    `json:"packageX86Url" gorm:"type:varchar(255)"`
	PackageARMURL   string                    `json:"packageArmUrl" gorm:"type:varchar(255)"`
	ChangelogZh     string                    `json:"changelogZh" gorm:"type:text"`
	ChangelogEn     string                    `json:"changelogEn" gorm:"type:text"`
	ReviewComment   string                    `json:"reviewComment" gorm:"type:text"`
	OfflineReasonZh string                    `json:"offlineReasonZh" gorm:"type:text"`
	OfflineReasonEn string                    `json:"offlineReasonEn" gorm:"type:text"`
	TDID            string                    `json:"tdId" gorm:"column:td_id;type:varchar(64)"`
	SubmittedAt     *time.Time                `json:"submittedAt"`
	ApprovedAt      *time.Time                `json:"approvedAt"`
	ReleasedAt      *time.Time                `json:"releasedAt"`
	OfflinedAt      *time.Time                `json:"offlinedAt"`
	CreatedBy       uint                      `json:"createdBy"`
	Plugin          Plugin                    `json:"plugin,omitempty" gorm:"foreignKey:PluginID"`
	Claimer         sysModel.SysUser          `json:"claimer,omitempty" gorm:"foreignKey:ClaimerID"`
	CompatibleItems []PluginCompatibleProduct `json:"compatibleItems,omitempty" gorm:"foreignKey:ReleaseID"`
}

func (PluginRelease) TableName() string { return "plugin_releases" }

type PluginCompatibleProduct struct {
	common.BaseModel
	ReleaseID         uint          `json:"releaseId" gorm:"uniqueIndex:idx_release_product;not null"`
	ProductID         uint          `json:"productId" gorm:"uniqueIndex:idx_release_product;not null"`
	Type              string        `json:"type" gorm:"type:varchar(16);uniqueIndex:idx_release_product;default:'product';not null"`
	VersionConstraint string        `json:"versionConstraint" gorm:"type:varchar(128)"`
	Product           PluginProduct `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

func (PluginCompatibleProduct) TableName() string { return "plugin_compatible_products" }

type PluginReleaseEvent struct {
	common.BaseModel
	ReleaseID         uint   `json:"releaseId" gorm:"index;not null"`
	FromStatus        int8   `json:"fromStatus"`
	ToStatus          int8   `json:"toStatus"`
	FromProcessStatus int8   `json:"fromProcessStatus"`
	ToProcessStatus   int8   `json:"toProcessStatus"`
	Action            string `json:"action" gorm:"type:varchar(32);index;not null"`
	OperatorID        uint   `json:"operatorId"`
	Comment           string `json:"comment" gorm:"type:text"`
	SnapshotJSON      string `json:"snapshotJson" gorm:"type:text"`
}

func (PluginReleaseEvent) TableName() string { return "plugin_release_events" }

func ToBaseModel(id uint) common.BaseModel {
	return common.BaseModel{ID: id}
}
