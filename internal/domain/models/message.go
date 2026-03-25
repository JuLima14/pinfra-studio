package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
	RoleSystem    = "system"
)

type Message struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	ChatID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Role       string    `gorm:"not null"`
	Content    string    `gorm:"type:text"`
	ToolName   string
	ToolInput  string `gorm:"type:jsonb"`
	ToolResult string `gorm:"type:jsonb"`
	CreatedAt  time.Time
	Chat Chat
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
