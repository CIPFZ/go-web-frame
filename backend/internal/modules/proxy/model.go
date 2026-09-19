package proxy

import "github.com/CIPFZ/gowebframe/internal/modules/common"

// Instance describes one locally managed proxy runtime. Secrets stay in the
// engine-owned configuration file and are never persisted in this table.
type Instance struct {
	common.BaseModel
	Name        string `json:"name" gorm:"size:100;uniqueIndex"`
	Engine      string `json:"engine" gorm:"size:20;index"`
	Scope       string `json:"scope" gorm:"size:20;default:system"`
	Unit        string `json:"unit" gorm:"size:128;index"`
	BinaryPath  string `json:"binaryPath" gorm:"size:255"`
	ConfigPath  string `json:"configPath" gorm:"size:255"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	Description string `json:"description" gorm:"size:255"`
}

func (Instance) TableName() string { return "proxy_instances" }
