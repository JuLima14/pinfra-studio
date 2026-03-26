package handlers

import (
	"github.com/JuLima14/pinfra-studio/internal/api/middleware"
	"github.com/JuLima14/pinfra-studio/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProjectHandler struct {
	projectService *services.ProjectService
	logger         *zap.Logger
}

func NewProjectHandler(projectService *services.ProjectService, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		logger:         logger,
	}
}

// ListProjects GET /api/v1/projects
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	projects, err := h.projectService.ListProjects(c.Context(), user.TenantID)
	if err != nil {
		h.logger.Error("list projects failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(projects)
}

// CreateProject POST /api/v1/projects
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	var body struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	project, err := h.projectService.CreateProject(c.Context(), user.TenantID, user.ID, body.Name)
	if err != nil {
		h.logger.Error("create project failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

// GetProject GET /api/v1/projects/:id
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := h.projectService.GetProject(c.Context(), user.TenantID, id)
	if err != nil {
		h.logger.Error("get project failed", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}
	return c.JSON(project)
}

// DeleteProject DELETE /api/v1/projects/:id
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	if err := h.projectService.DeleteProject(c.Context(), user.TenantID, id); err != nil {
		h.logger.Error("delete project failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
