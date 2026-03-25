package repositories

import (
	"context"
	"fmt"

	"github.com/JuLima14/pinfra-studio/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	FindByChat(ctx context.Context, chatID uuid.UUID) ([]*models.Message, error)
	DeleteByChat(ctx context.Context, chatID uuid.UUID) error
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	if err := r.db.WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (r *messageRepository) FindByChat(ctx context.Context, chatID uuid.UUID) ([]*models.Message, error) {
	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to find messages by chat: %w", err)
	}
	return messages, nil
}

func (r *messageRepository) DeleteByChat(ctx context.Context, chatID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Message{}, "chat_id = ?", chatID).Error; err != nil {
		return fmt.Errorf("failed to delete messages by chat: %w", err)
	}
	return nil
}
