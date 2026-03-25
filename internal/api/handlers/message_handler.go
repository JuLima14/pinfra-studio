package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"time"

	"github.com/JuLima14/claude-engine/events"
	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/JuLima14/pinfra-studio/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MessageHandler struct {
	chatService *services.ChatService
	projectRepo repositories.ProjectRepository
	broadcaster *events.Broadcaster
	logger      *zap.Logger
}

func NewMessageHandler(
	chatService *services.ChatService,
	projectRepo repositories.ProjectRepository,
	broadcaster *events.Broadcaster,
	logger *zap.Logger,
) *MessageHandler {
	return &MessageHandler{
		chatService: chatService,
		projectRepo: projectRepo,
		broadcaster: broadcaster,
		logger:      logger,
	}
}

// SendMessage POST /api/v1/projects/:id/chats/:chatId/messages
func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "content is required"})
	}

	if err := h.chatService.SendMessage(c.Context(), chatID, body.Content); err != nil {
		h.logger.Error("send message failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "processing"})
}

// CancelMessage POST /api/v1/projects/:id/chats/:chatId/messages/cancel
func (h *MessageHandler) CancelMessage(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	h.chatService.CancelGeneration(projectID)
	return c.JSON(fiber.Map{"status": "cancelled"})
}

// StreamEvents GET /api/v1/projects/:id/chats/:chatId/events
// SSE endpoint — streams events for the given chat channel.
func (h *MessageHandler) StreamEvents(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	channelID := chatID.String()

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	// chanWriter bridges broadcaster (net/http.ResponseWriter) with Fiber's stream writer.
	// We use a buffered channel so Broadcast calls don't block.
	dataCh := make(chan string, 64)

	// Create a net/http adapter that writes to dataCh.
	adapter := newChannelResponseWriter(dataCh)
	client := h.broadcaster.Register(channelID, adapter)
	if client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to register SSE client"})
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer h.broadcaster.Remove(channelID, client)

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg, ok := <-dataCh:
				if !ok {
					return
				}
				if _, err := fmt.Fprint(w, msg); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}

// channelResponseWriter implements net/http.ResponseWriter + net/http.Flusher.
// It forwards written SSE data to an internal channel instead of a real connection,
// allowing the broadcaster (which expects net/http) to work with Fiber's stream writer.
type channelResponseWriter struct {
	ch     chan string
	header http.Header
}

func newChannelResponseWriter(ch chan string) *channelResponseWriter {
	return &channelResponseWriter{
		ch:     ch,
		header: make(http.Header),
	}
}

func (w *channelResponseWriter) Header() http.Header {
	return w.header
}

func (w *channelResponseWriter) Write(p []byte) (int, error) {
	// Non-blocking send: drop if channel is full to avoid blocking broadcaster.
	select {
	case w.ch <- string(p):
	default:
	}
	return len(p), nil
}

func (w *channelResponseWriter) WriteHeader(statusCode int) {}

// Flush implements http.Flusher — broadcaster calls this after each write.
// The actual flush happens in the stream writer goroutine.
func (w *channelResponseWriter) Flush() {}

