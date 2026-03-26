package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	SandboxStarting = "starting"
	SandboxRunning  = "running"
	SandboxStopped  = "stopped"
	SandboxError    = "error"
)

type Sandbox struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"projectId"`
	ContainerID  string    `json:"-"`
	Port         int       `json:"port"`
	Status       string    `gorm:"default:'stopped'" json:"status"`
	LastActiveAt time.Time `json:"lastActiveAt"`
	CreatedAt    time.Time `json:"createdAt"`
	Project      Project   `json:"-"`
}

func (s *Sandbox) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
