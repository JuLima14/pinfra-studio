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
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index" json:"tenantId"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Name          string    `gorm:"not null" json:"name"`
	Slug          string    `gorm:"uniqueIndex;not null" json:"slug"`
	Template      string    `gorm:"default:'next-app'" json:"template"`
	GitHubRepoURL string    `json:"githubRepoUrl,omitempty"`
	GitHubBranch  string    `gorm:"default:'main'" json:"githubBranch"`
	Status        string    `gorm:"default:'active'" json:"status"`
	SetupStatus   string    `gorm:"default:'scaffolding'" json:"setupStatus"`
	SetupError    string    `json:"setupError,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Chats         []Chat    `json:"chats,omitempty"`
	Sandbox       *Sandbox  `json:"sandbox,omitempty"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
