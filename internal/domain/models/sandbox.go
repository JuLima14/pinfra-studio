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
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	ContainerID  string
	Port         int
	Status       string `gorm:"default:'stopped'"`
	LastActiveAt time.Time
	CreatedAt    time.Time
	Project Project
}

func (s *Sandbox) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
