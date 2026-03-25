package repositories

import (
	"context"
	"fmt"

	"github.com/JuLima14/pinfra-studio/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat) error
	Update(ctx context.Context, chat *models.Chat) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Chat, error)
	FindByProject(ctx context.Context, projectID uuid.UUID) ([]*models.Chat, error)
	FindActiveByProject(ctx context.Context, projectID uuid.UUID) (*models.Chat, error)
	SetActive(ctx context.Context, projectID uuid.UUID, chatID uuid.UUID) error
	CountByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(ctx context.Context, chat *models.Chat) error {
	if err := r.db.WithContext(ctx).Create(chat).Error; err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}
	return nil
}

func (r *chatRepository) Update(ctx context.Context, chat *models.Chat) error {
	if err := r.db.WithContext(ctx).Save(chat).Error; err != nil {
		return fmt.Errorf("failed to update chat: %w", err)
	}
	return nil
}

func (r *chatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Message{}, "chat_id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to cascade delete messages: %w", err)
	}
	if err := r.db.WithContext(ctx).Delete(&models.Chat{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}
	return nil
}

func (r *chatRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	if err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ?", id).
		First(&chat).Error; err != nil {
		return nil, fmt.Errorf("failed to find chat by id: %w", err)
	}
	return &chat, nil
}

func (r *chatRepository) FindByProject(ctx context.Context, projectID uuid.UUID) ([]*models.Chat, error) {
	var chats []*models.Chat
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&chats).Error; err != nil {
		return nil, fmt.Errorf("failed to find chats by project: %w", err)
	}
	return chats, nil
}

func (r *chatRepository) FindActiveByProject(ctx context.Context, projectID uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND is_active = true", projectID).
		First(&chat).Error; err != nil {
		return nil, fmt.Errorf("failed to find active chat by project: %w", err)
	}
	return &chat, nil
}

func (r *chatRepository) SetActive(ctx context.Context, projectID uuid.UUID, chatID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Chat{}).
			Where("project_id = ?", projectID).
			Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate chats: %w", err)
		}
		if err := tx.Model(&models.Chat{}).
			Where("id = ? AND project_id = ?", chatID, projectID).
			Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate chat: %w", err)
		}
		return nil
	})
}

func (r *chatRepository) CountByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Chat{}).
		Where("project_id = ?", projectID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count chats by project: %w", err)
	}
	return count, nil
}
