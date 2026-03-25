package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Chat struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Title           string
	BranchName      string `gorm:"not null"`
	ClaudeSessionID string
	Status          string `gorm:"default:'active'"`
	IsActive        bool   `gorm:"default:false"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Project  Project
	Messages []Message
}

func (c *Chat) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
