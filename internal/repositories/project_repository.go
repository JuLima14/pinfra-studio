package repositories

import (
	"context"
	"fmt"

	"github.com/JuLima14/pinfra-studio/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Project, error)
	FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.Project, error)
	UpdateSetupStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	if err := r.db.WithContext(ctx).Save(project).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Project{}, "id = ? AND tenant_id = ?", id, tenantID).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

func (r *projectRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).
		Preload("Chats").
		Preload("Sandbox").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project by id: %w", err)
	}
	return &project, nil
}

func (r *projectRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.Project, error) {
	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Sandbox").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to find projects by tenant: %w", err)
	}
	return projects, nil
}

func (r *projectRepository) UpdateSetupStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"setup_status": status,
		"setup_error":  errorMsg,
	}
	if err := r.db.WithContext(ctx).Model(&models.Project{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update setup status: %w", err)
	}
	return nil
}
