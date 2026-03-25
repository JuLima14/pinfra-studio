package handlers

import (
	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/JuLima14/pinfra-studio/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SandboxHandler struct {
	sandboxService *services.SandboxService
	projectRepo    repositories.ProjectRepository
	logger         *zap.Logger
}

func NewSandboxHandler(
	sandboxService *services.SandboxService,
	projectRepo repositories.ProjectRepository,
	logger *zap.Logger,
) *SandboxHandler {
	return &SandboxHandler{
		sandboxService: sandboxService,
		projectRepo:    projectRepo,
		logger:         logger,
	}
}

// GetStatus GET /api/v1/projects/:id/sandbox/status
func (h *SandboxHandler) GetStatus(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	sbx, err := h.sandboxService.GetStatus(c.Context(), projectID)
	if err != nil {
		h.logger.Error("get sandbox status failed", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(sbx)
}

// StartSandbox POST /api/v1/projects/:id/sandbox/start
func (h *SandboxHandler) StartSandbox(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	if err := h.sandboxService.Start(c.Context(), projectID); err != nil {
		h.logger.Error("start sandbox failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "started"})
}

// StopSandbox POST /api/v1/projects/:id/sandbox/stop
func (h *SandboxHandler) StopSandbox(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	if err := h.sandboxService.Stop(c.Context(), projectID); err != nil {
		h.logger.Error("stop sandbox failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "stopped"})
}

// GetURL GET /api/v1/projects/:id/sandbox/url
func (h *SandboxHandler) GetURL(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	url, err := h.sandboxService.GetURL(c.Context(), projectID)
	if err != nil {
		h.logger.Error("get sandbox url failed", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"url": url})
}

// GetLogs GET /api/v1/projects/:id/sandbox/logs
func (h *SandboxHandler) GetLogs(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	logs, err := h.sandboxService.GetLogs(c.Context(), projectID, 100)
	if err != nil {
		h.logger.Error("get sandbox logs failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"logs": logs})
}
