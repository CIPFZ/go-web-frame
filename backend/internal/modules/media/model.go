package media

import (
	"time"
)

// Base exposes the lower-camel API contract without affecting the legacy CMS models.
type Base struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Source is a reviewed upstream MacCMS-compatible API.
type Source struct {
	Base
	Suspended     bool       `json:"suspended"`
	Failures      int        `json:"failures"`
	Successes     int        `json:"successes"`
	LeaseToken    string     `json:"-" gorm:"size:64"`
	LeaseUntil    *time.Time `json:"-"`
	FullStartedAt *time.Time `json:"fullStartedAt"`
	NextPage      int        `json:"nextPage"`

	Name                   string     `json:"name" gorm:"size:100;not null"`
	APIURL                 string     `json:"apiUrl" gorm:"size:500;uniqueIndex:uk_media_source_api;not null"`
	APIType                string     `json:"apiType" gorm:"size:10;default:json;not null"`
	Enabled                bool       `json:"enabled" gorm:"default:true;index"`
	Priority               int        `json:"priority" gorm:"default:100;index"`
	AllowCategories        string     `json:"allowCategories,omitempty" gorm:"type:text"`
	DenyCategories         string     `json:"denyCategories,omitempty" gorm:"type:text"`
	RequestIntervalSeconds int        `json:"requestIntervalSeconds" gorm:"default:3"`
	TimeoutSeconds         int        `json:"timeoutSeconds" gorm:"default:15"`
	LastOKAt               *time.Time `json:"lastOkAt"`
	LastErrorAt            *time.Time `json:"lastErrorAt"`
	LastError              string     `json:"lastError,omitempty" gorm:"type:text"`
}

func (Source) TableName() string { return "media_sources" }

// Item is the unified CMS media record. The host CMS can map CMSMediaID later.
type Item struct {
	Base
	CMSMediaID    *uint  `json:"cmsMediaId"`
	Title         string `json:"title" gorm:"size:500;not null"`
	OriginalTitle string `json:"originalTitle,omitempty" gorm:"size:500"`
	MediaType     string `json:"mediaType" gorm:"size:20;index;not null"`
	Year          string `json:"year,omitempty" gorm:"size:20;index"`
	Area          string `json:"area,omitempty" gorm:"size:100"`
	Language      string `json:"language,omitempty" gorm:"size:100"`
	Director      string `json:"director,omitempty" gorm:"type:text"`
	Actors        string `json:"actors,omitempty" gorm:"type:text"`
	Description   string `json:"description,omitempty" gorm:"type:text"`
	CoverURL      string `json:"coverUrl,omitempty" gorm:"size:1000"`
	NormalizedKey string `json:"normalizedKey" gorm:"size:700;uniqueIndex:uk_media_item_key;not null"`
	Status        string `json:"status" gorm:"size:20;default:active;index"`
}

func (Item) TableName() string { return "media_items" }

// SourceItem preserves the upstream identity and raw metadata for reconciliation.
type SourceItem struct {
	Base
	SourceID        uint       `json:"sourceId" gorm:"index;not null;uniqueIndex:uk_source_external"`
	CMSMediaID      *uint      `json:"cmsMediaId"`
	ExternalID      string     `json:"externalId" gorm:"size:100;not null;uniqueIndex:uk_source_external"`
	ExternalTitle   string     `json:"externalTitle" gorm:"size:500;not null"`
	SourceUpdatedAt *time.Time `json:"sourceUpdatedAt"`
	FirstSeenAt     time.Time  `json:"firstSeenAt"`
	LastSeenAt      time.Time  `json:"lastSeenAt"`
	RawHash         string     `json:"rawHash,omitempty" gorm:"size:64"`
	DetailURL       string     `json:"detailUrl,omitempty" gorm:"size:1000"`
	State           string     `json:"state" gorm:"size:20;default:active;index"`
	RawMetadata     string     `json:"rawMetadata,omitempty" gorm:"type:text"`
}

func (SourceItem) TableName() string { return "media_source_items" }

// PlayLine belongs to one source item; source lines are never merged across sources.
type PlayLine struct {
	Base
	Episodes   []PlayEpisode `json:"episodes" gorm:"foreignKey:PlayLineID"`
	SourceName string        `json:"sourceName" gorm:"-"`

	SourceItemID  uint       `json:"sourceItemId" gorm:"index;not null;uniqueIndex:uk_source_media_line"`
	LineName      string     `json:"lineName" gorm:"size:100;not null"`
	LineKey       string     `json:"lineKey" gorm:"size:100;not null;uniqueIndex:uk_source_media_line"`
	SortOrder     int        `json:"sortOrder" gorm:"default:100;index"`
	Enabled       bool       `json:"enabled" gorm:"default:true;index"`
	LastCheckedAt *time.Time `json:"lastCheckedAt"`
	LastOKAt      *time.Time `json:"lastOkAt"`
	LastLatencyMs *int       `json:"lastLatencyMs"`
	LastStatus    string     `json:"lastStatus,omitempty" gorm:"size:30"`
}

func (PlayLine) TableName() string { return "media_play_lines" }

type PlayEpisode struct {
	Base
	PlayLineID    uint       `json:"playLineId" gorm:"index;not null;uniqueIndex:uk_line_episode"`
	EpisodeNo     int        `json:"episodeNo" gorm:"not null;uniqueIndex:uk_line_episode"`
	EpisodeName   string     `json:"episodeName,omitempty" gorm:"size:255"`
	StreamURL     string     `json:"streamUrl" gorm:"type:text;not null"`
	StreamType    string     `json:"streamType" gorm:"size:20;default:m3u8"`
	URLHash       string     `json:"urlHash" gorm:"size:64;index"`
	LastCheckedAt *time.Time `json:"lastCheckedAt"`
	LastOKAt      *time.Time `json:"lastOkAt"`
	LastStatus    string     `json:"lastStatus,omitempty" gorm:"size:30"`
	ExpiresAt     *time.Time `json:"expiresAt"`
}

func (PlayEpisode) TableName() string { return "media_play_episodes" }

type CollectionRun struct {
	Base
	NextPage int `json:"nextPage"`

	SourceID      uint       `json:"sourceId" gorm:"index;not null"`
	Mode          string     `json:"mode" gorm:"size:20;not null"`
	StartedAt     time.Time  `json:"startedAt" gorm:"index;not null"`
	FinishedAt    *time.Time `json:"finishedAt"`
	Status        string     `json:"status" gorm:"size:20;index;not null"`
	Pages         int        `json:"pages"`
	Fetched       int        `json:"fetched"`
	InsertedCount int        `json:"insertedCount"`
	UpdatedCount  int        `json:"updatedCount"`
	SkippedCount  int        `json:"skippedCount"`
	ErrorCount    int        `json:"errorCount"`
	ErrorMessage  string     `json:"errorMessage,omitempty" gorm:"type:text"`
}

func (CollectionRun) TableName() string { return "media_collection_runs" }

type HealthCheck struct {
	Base
	SourceID      uint      `json:"sourceId" gorm:"index;not null"`
	CheckedAt     time.Time `json:"checkedAt" gorm:"index;not null"`
	HTTPStatus    int       `json:"httpStatus"`
	LatencyMS     int       `json:"latencyMs"`
	ResponseBytes int       `json:"responseBytes"`
	Result        string    `json:"result" gorm:"size:20;index;not null"`
	ErrorMessage  string    `json:"errorMessage,omitempty" gorm:"size:500"`
}

func (HealthCheck) TableName() string { return "media_source_health_checks" }
