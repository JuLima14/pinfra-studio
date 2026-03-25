package handlers

import (
	"path/filepath"

	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/JuLima14/pinfra-studio/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GitHubHandler struct {
	githubService *services.GitHubService
	gitService    *services.GitService
	projectRepo   repositories.ProjectRepository
	dataDir       string
	logger        *zap.Logger
}

func NewGitHubHandler(
	githubService *services.GitHubService,
	gitService *services.GitService,
	projectRepo repositories.ProjectRepository,
	dataDir string,
	logger *zap.Logger,
) *GitHubHandler {
	return &GitHubHandler{
		githubService: githubService,
		gitService:    gitService,
		projectRepo:   projectRepo,
		dataDir:       dataDir,
		logger:        logger,
	}
}

// ConnectGitHub POST /api/v1/projects/:id/github/connect
func (h *GitHubHandler) ConnectGitHub(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	var body struct {
		RepoURL string `json:"repo_url"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.RepoURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "repo_url is required"})
	}

	project, err := h.projectRepo.FindByID(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	projectDir := filepath.Join(h.dataDir, project.ID.String())
	if err := h.githubService.ConnectRepo(projectDir, body.RepoURL); err != nil {
		h.logger.Error("connect github repo failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Persist the GitHub repo URL
	project.GitHubRepoURL = body.RepoURL
	if err := h.projectRepo.Update(c.Context(), project); err != nil {
		h.logger.Warn("failed to persist github repo url", zap.Error(err))
	}

	return c.JSON(fiber.Map{"status": "connected", "repo_url": body.RepoURL})
}

// PushToGitHub POST /api/v1/projects/:id/github/push
func (h *GitHubHandler) PushToGitHub(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := h.projectRepo.FindByID(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	projectDir := filepath.Join(h.dataDir, project.ID.String())

	branch, err := h.gitService.CurrentBranch(projectDir)
	if err != nil {
		h.logger.Error("get current branch failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.githubService.Push(projectDir, branch); err != nil {
		h.logger.Error("push to github failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "pushed", "branch": branch})
}

// CreatePR POST /api/v1/projects/:id/github/pr
func (h *GitHubHandler) CreatePR(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	var body struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
	}

	project, err := h.projectRepo.FindByID(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	projectDir := filepath.Join(h.dataDir, project.ID.String())

	prURL, err := h.githubService.CreatePR(projectDir, body.Title, body.Body)
	if err != nil {
		h.logger.Error("create PR failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"url": prURL})
}
