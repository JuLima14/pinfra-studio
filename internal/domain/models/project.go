package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	SetupScaffolding = "scaffolding"
	SetupInstalling  = "installing"
	SetupStarting    = "starting"
	SetupReady       = "ready"
	SetupFailed      = "failed"
)

type Project struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name          string    `gorm:"not null"`
	Slug          string    `gorm:"uniqueIndex;not null"`
	Template      string    `gorm:"default:'next-app'"`
	GitHubRepoURL string
	GitHubBranch  string `gorm:"default:'main'"`
	Status        string `gorm:"default:'active'"`
	SetupStatus   string `gorm:"default:'scaffolding'"`
	SetupError    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Chats   []Chat
	Sandbox *Sandbox
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
