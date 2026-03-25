package handlers

import (
	"github.com/JuLima14/pinfra-studio/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChatHandler struct {
	chatService *services.ChatService
	logger      *zap.Logger
}

func NewChatHandler(chatService *services.ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// ListChats GET /api/v1/projects/:id/chats
func (h *ChatHandler) ListChats(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	chats, err := h.chatService.ListChats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("list chats failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(chats)
}

// CreateChat POST /api/v1/projects/:id/chats
func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	chat, err := h.chatService.CreateChat(c.Context(), projectID)
	if err != nil {
		h.logger.Error("create chat failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(chat)
}

// GetChat GET /api/v1/projects/:id/chats/:chatId
func (h *ChatHandler) GetChat(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	chat, err := h.chatService.GetChat(c.Context(), chatID)
	if err != nil {
		h.logger.Error("get chat failed", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "chat not found"})
	}
	return c.JSON(chat)
}

// ActivateChat POST /api/v1/projects/:id/chats/:chatId/activate
func (h *ChatHandler) ActivateChat(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	if err := h.chatService.ActivateChat(c.Context(), projectID, chatID); err != nil {
		h.logger.Error("activate chat failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

// DeleteChat DELETE /api/v1/projects/:id/chats/:chatId
func (h *ChatHandler) DeleteChat(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	if err := h.chatService.DeleteChat(c.Context(), chatID); err != nil {
		h.logger.Error("delete chat failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
