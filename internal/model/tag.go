package model

import (
	"KnowLedger/pkg/utils"

	"gorm.io/gorm"
)

type Tag struct {
	ID   string `gorm:"primary_key"`
	Name string `gorm:"type:varchar(100)"`

	Facts []Fact `gorm:"many2many:fact_tags" json:"Facts,omitempty"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = utils.GenerateRandomULID()
	}

	t.Name = utils.NormalizeTagName(t.Name)

	return nil
}

func (t *Tag) BeforeUpdate(tx *gorm.DB) error {
	t.Name = utils.NormalizeTagName(t.Name)
	return nil
}
