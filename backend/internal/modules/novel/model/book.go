package model

import (
	"time"

	"gorm.io/datatypes"
)

// NovelBook is a book record imported from the ebook_treasure_chest catalog.
// The source catalog remains the system of record for future imports; this
// table is the CMS-local copy used by the novel module.
type NovelBook struct {
	ID         uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Title      string         `json:"title" gorm:"size:500;not null;index:idx_novel_book_title,priority:1"`
	Author     string         `json:"author" gorm:"size:500;index:idx_novel_book_author,priority:1"`
	Category   string         `json:"category" gorm:"size:200;index:idx_novel_book_category,priority:1"`
	SourceURL  string         `json:"sourceUrl" gorm:"type:text;not null"`
	SourceHash string         `json:"sourceHash" gorm:"type:char(64);not null;index:idx_novel_book_source_hash"`
	RecordHash string         `json:"recordHash" gorm:"type:char(64);not null;uniqueIndex:uk_novel_book_record_hash"`
	Language   string         `json:"language" gorm:"size:20"`
	Formats    datatypes.JSON `json:"formats" gorm:"type:json"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

func (NovelBook) TableName() string { return "novel_books" }
