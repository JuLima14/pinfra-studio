package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is a read-only view of the infra-platform users table.
// pinfra-studio shares the same PostgreSQL database.
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenantId"`
	Auth0ID   string         `gorm:"type:varchar(255);unique;not null" json:"auth0Id"`
	Email     string         `gorm:"type:varchar(255);not null" json:"email"`
	Name      string         `gorm:"type:varchar(255)" json:"name"`
	Role      string         `gorm:"type:varchar(50);not null;default:'viewer'" json:"role"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Tenant is a read-only view of the infra-platform tenants table.
type Tenant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug      string         `gorm:"type:varchar(100);unique;not null" json:"slug"`
	Type      string         `gorm:"type:varchar(50);not null;default:'user'" json:"type"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
