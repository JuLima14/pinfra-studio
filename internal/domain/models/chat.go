package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Chat struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;index" json:"projectId"`
	Title           string    `json:"title,omitempty"`
	BranchName      string    `gorm:"not null" json:"branchName"`
	ClaudeSessionID string    `json:"-"`
	Status          string    `gorm:"default:'active'" json:"status"`
	IsActive        bool      `gorm:"default:false" json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Project         Project   `json:"-"`
	Messages        []Message `json:"messages,omitempty"`
}

func (c *Chat) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
