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
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ChatID     uuid.UUID `gorm:"type:uuid;not null;index" json:"chatId"`
	Role       string    `gorm:"not null" json:"role"`
	Content    string    `gorm:"type:text" json:"content"`
	ToolName   string  `json:"toolName,omitempty"`
	ToolInput  *string `gorm:"type:jsonb" json:"toolInput,omitempty"`
	ToolResult *string `gorm:"type:jsonb" json:"toolResult,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	Chat       Chat      `json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
