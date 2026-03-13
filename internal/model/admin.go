package model

import (
	"KnowLedger/pkg/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	ID        string `gorm:"primary_key"`
	Username  string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	ApiKey    string `gorm:"uniqueIndex;default:''"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Admin) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = utils.GenerateRandomULID()
	}
	m.Username = strings.TrimSpace(strings.ToLower(m.Username))

	if m.ApiKey == "" {
		apikey, err := utils.GenerateAPIKey(32, "knowledger_j")
		if err != nil {
			apikey = ""
		}
		m.ApiKey = apikey
	}

	return nil
}
