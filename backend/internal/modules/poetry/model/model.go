package model

import "time"

type PoetryBase struct {
	ID        uint      `gorm:"primarykey" json:"ID"` // 主键
	CreatedAt time.Time `json:"createdAt"`            // 创建时间
	UpdatedAt time.Time `json:"updatedAt"`            // 更新时间
}

// MetaDynasty 朝代
type MetaDynasty struct {
	PoetryBase        // ✨ 替换 common.BaseModel
	Name       string `json:"name" gorm:"type:varchar(50);uniqueIndex;not null"`
	SortOrder  int    `json:"sortOrder" gorm:"default:0"`
}

func (MetaDynasty) TableName() string { return "meta_dynasty" }

// MetaGenre 体裁
type MetaGenre struct {
	PoetryBase        // ✨ 替换 common.BaseModel
	Name       string `json:"name" gorm:"type:varchar(20);uniqueIndex;not null"`
	SortOrder  int    `json:"sortOrder" gorm:"default:0"`
}

func (MetaGenre) TableName() string { return "meta_genre" }

// MetaTag 标签
type MetaTag struct {
	PoetryBase        // ✨ 替换 common.BaseModel
	Name       string `json:"name" gorm:"type:varchar(50);uniqueIndex;not null"`
	Category   string `json:"category" gorm:"type:varchar(20)"`
	SortOrder  int    `json:"sortOrder" gorm:"default:0"`
}

func (MetaTag) TableName() string { return "meta_tag" }

// PoemAuthor 诗人
type PoemAuthor struct {
	PoetryBase        // ✨ 替换 common.BaseModel
	Name       string `json:"name" gorm:"type:varchar(100);index;not null"`
	DynastyID  uint   `json:"dynastyId" gorm:"not null;index"`
	Intro      string `json:"intro" gorm:"type:text"`
	LifeStory  string `json:"lifeStory" gorm:"type:text"`
	AvatarUrl  string `json:"avatarUrl" gorm:"type:varchar(500)"`

	Dynasty MetaDynasty `json:"dynasty" gorm:"foreignKey:DynastyID"`
}

func (PoemAuthor) TableName() string { return "poem_author" }

// PoemWork 诗词作品
type PoemWork struct {
	PoetryBase          // ✨ 替换 common.BaseModel
	Title        string `json:"title" gorm:"type:varchar(200);index;not null"`
	AuthorID     uint   `json:"authorId" gorm:"not null;index"`
	GenreID      uint   `json:"genreId" gorm:"not null;index"`
	Content      string `json:"content" gorm:"type:text;not null"`
	Translation  string `json:"translation" gorm:"type:text"`
	Annotation   string `json:"annotation" gorm:"type:text"`
	Appreciation string `json:"appreciation" gorm:"type:text"`
	AudioUrl     string `json:"audioUrl" gorm:"type:varchar(500)"`
	ViewCount    int    `json:"viewCount" gorm:"default:0"`

	Author PoemAuthor `json:"author" gorm:"foreignKey:AuthorID"`
	Genre  MetaGenre  `json:"genre" gorm:"foreignKey:GenreID"`
	Tags   []MetaTag  `json:"tags" gorm:"many2many:poem_tag_rel;joinForeignKey:work_id;joinReferences:tag_id"`
}

func (PoemWork) TableName() string { return "poem_work" }

// PoemTagRel 关联表
type PoemTagRel struct {
	WorkID uint `gorm:"primaryKey;column:work_id;type:bigint unsigned;not null"`
	TagID  uint `gorm:"primaryKey;column:tag_id;type:bigint unsigned;not null"`
}

func (PoemTagRel) TableName() string { return "poem_tag_rel" }
