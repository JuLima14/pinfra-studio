package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/JuLima14/pinfra-studio/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SandboxRepository interface {
	Create(ctx context.Context, sandbox *models.Sandbox) error
	Update(ctx context.Context, sandbox *models.Sandbox) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByProject(ctx context.Context, projectID uuid.UUID) (*models.Sandbox, error)
	FindAll(ctx context.Context) ([]*models.Sandbox, error)
	FindStale(ctx context.Context, idleTimeout time.Duration) ([]*models.Sandbox, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	TouchLastActive(ctx context.Context, id uuid.UUID) error
}

type sandboxRepository struct {
	db *gorm.DB
}

func NewSandboxRepository(db *gorm.DB) SandboxRepository {
	return &sandboxRepository{db: db}
}

func (r *sandboxRepository) Create(ctx context.Context, sandbox *models.Sandbox) error {
	if err := r.db.WithContext(ctx).Create(sandbox).Error; err != nil {
		return fmt.Errorf("failed to create sandbox: %w", err)
	}
	return nil
}

func (r *sandboxRepository) Update(ctx context.Context, sandbox *models.Sandbox) error {
	if err := r.db.WithContext(ctx).Save(sandbox).Error; err != nil {
		return fmt.Errorf("failed to update sandbox: %w", err)
	}
	return nil
}

func (r *sandboxRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Sandbox{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete sandbox: %w", err)
	}
	return nil
}

func (r *sandboxRepository) FindByProject(ctx context.Context, projectID uuid.UUID) (*models.Sandbox, error) {
	var sandbox models.Sandbox
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		First(&sandbox).Error; err != nil {
		return nil, fmt.Errorf("failed to find sandbox by project: %w", err)
	}
	return &sandbox, nil
}

func (r *sandboxRepository) FindAll(ctx context.Context) ([]*models.Sandbox, error) {
	var sandboxes []*models.Sandbox
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Find(&sandboxes).Error; err != nil {
		return nil, fmt.Errorf("failed to find all sandboxes: %w", err)
	}
	return sandboxes, nil
}

func (r *sandboxRepository) FindStale(ctx context.Context, idleTimeout time.Duration) ([]*models.Sandbox, error) {
	cutoff := time.Now().Add(-idleTimeout)
	var sandboxes []*models.Sandbox
	if err := r.db.WithContext(ctx).
		Where("status = ? AND last_active_at < ?", models.SandboxRunning, cutoff).
		Find(&sandboxes).Error; err != nil {
		return nil, fmt.Errorf("failed to find stale sandboxes: %w", err)
	}
	return sandboxes, nil
}

func (r *sandboxRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.Sandbox{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update sandbox status: %w", err)
	}
	return nil
}

func (r *sandboxRepository) TouchLastActive(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Sandbox{}).
		Where("id = ?", id).
		Update("last_active_at", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to touch last active: %w", err)
	}
	return nil
}
