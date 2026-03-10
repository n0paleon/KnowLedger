package model

import (
	"KnowLedger/pkg/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type FactStatus string

const (
	FactStatusDraft     FactStatus = "draft"
	FactStatusPublished FactStatus = "published"
)

type Fact struct {
	ID      string     `gorm:"primary_key"`
	Content string     `gorm:"notnull;type:text"`
	Status  FactStatus `gorm:"notnull;default:'draft'"`

	SourceURL string `gorm:"type:text"`

	Media *MediaItem `gorm:"type:jsonb;default:null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Tags []Tag `gorm:"many2many:fact_tags"`
}

func (f *Fact) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = utils.GenerateRandomULID()
	}
	f.Content = strings.TrimSpace(f.Content)

	return nil
}

func (f *Fact) GetTagsString() string {
	names := make([]string, len(f.Tags))
	for i, t := range f.Tags {
		names[i] = t.Name
	}
	return strings.Join(names, ", ")
}

type FunFactStats struct {
	FunFacts int64
	Tags     int64
}

// FactTag is Join table model
type FactTag struct {
	FactID string `gorm:"primaryKey"`
	TagID  string `gorm:"primaryKey"`
}
