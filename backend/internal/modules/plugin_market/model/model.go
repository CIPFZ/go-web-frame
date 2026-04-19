package model

import "time"

type MarketPlugin struct {
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PluginID      uint   `gorm:"uniqueIndex;not null"`
	Code          string `gorm:"type:varchar(64);index;not null"`
	NameZh        string `gorm:"type:varchar(128);not null"`
	NameEn        string `gorm:"type:varchar(128);not null"`
	DescriptionZh string `gorm:"type:text;not null"`
	DescriptionEn string `gorm:"type:text;not null"`
	CapabilityZh  string `gorm:"type:text"`
	CapabilityEn  string `gorm:"type:text"`
	OwnerName     string `gorm:"type:varchar(128)"`
	Status        string `gorm:"type:varchar(16);index;not null;default:'published'"`
	LatestVersion string `gorm:"type:varchar(64)"`

	Versions []MarketPluginVersion `gorm:"foreignKey:PluginRefID"`
}

func (MarketPlugin) TableName() string { return "market_plugins" }

type MarketPluginVersion struct {
	ID                 uint `gorm:"primarykey"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	PluginRefID        uint   `gorm:"index;not null"`
	PluginID           uint   `gorm:"index;not null"`
	ReleaseID          uint   `gorm:"uniqueIndex;not null"`
	Version            string `gorm:"type:varchar(64);not null"`
	ChangelogZh        string `gorm:"type:text"`
	ChangelogEn        string `gorm:"type:text"`
	TestReportURL      string `gorm:"type:varchar(255)"`
	PackageX86URL      string `gorm:"type:varchar(255)"`
	PackageARMURL      string `gorm:"type:varchar(255)"`
	ReleasedAt         *time.Time
	VersionConstraint  string `gorm:"type:varchar(128)"`
	Publisher          string `gorm:"type:varchar(128)"`
	Status             string `gorm:"type:varchar(16);index;not null;default:'published'"`
	PluginNameZh       string `gorm:"type:varchar(128)"`
	PluginNameEn       string `gorm:"type:varchar(128)"`
	DescriptionZh      string `gorm:"type:text"`
	DescriptionEn      string `gorm:"type:text"`
	CapabilityZh       string `gorm:"type:text"`
	CapabilityEn       string `gorm:"type:text"`
	OwnerName          string `gorm:"type:varchar(128)"`
	CompatibleItems    []MarketPluginCompatibility `gorm:"foreignKey:VersionRefID"`
}

func (MarketPluginVersion) TableName() string { return "market_plugin_versions" }

type MarketPluginCompatibility struct {
	ID                uint `gorm:"primarykey"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	VersionRefID      uint   `gorm:"index;not null"`
	ReleaseID         uint   `gorm:"index;not null"`
	TargetType        string `gorm:"type:varchar(16);index;not null"`
	ProductCode       string `gorm:"type:varchar(64);not null"`
	ProductName       string `gorm:"type:varchar(128);not null"`
	VersionConstraint string `gorm:"type:varchar(128)"`
}

func (MarketPluginCompatibility) TableName() string { return "market_plugin_compatibilities" }
