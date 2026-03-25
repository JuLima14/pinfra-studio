package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/JuLima14/pinfra-studio/internal/domain/models"
	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/JuLima14/pinfra-studio/internal/sandbox"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SetupEventFunc func(projectID uuid.UUID, status, message string)

type ProjectService struct {
	projectRepo  repositories.ProjectRepository
	chatRepo     repositories.ChatRepository
	sandboxRepo  repositories.SandboxRepository
	sandboxMgr   *sandbox.Manager
	gitService   *GitService
	logger       *zap.Logger
	dataDir      string
	onSetupEvent SetupEventFunc
}

func NewProjectService(
	projectRepo repositories.ProjectRepository,
	chatRepo repositories.ChatRepository,
	sandboxRepo repositories.SandboxRepository,
	sandboxMgr *sandbox.Manager,
	gitService *GitService,
	logger *zap.Logger,
	dataDir string,
	onSetupEvent SetupEventFunc,
) *ProjectService {
	return &ProjectService{
		projectRepo:  projectRepo,
		chatRepo:     chatRepo,
		sandboxRepo:  sandboxRepo,
		sandboxMgr:   sandboxMgr,
		gitService:   gitService,
		logger:       logger,
		dataDir:      dataDir,
		onSetupEvent: onSetupEvent,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, name string) (*models.Project, error) {
	slug := generateSlug(name)

	project := &models.Project{
		Name:        name,
		Slug:        slug,
		Template:    "next-app",
		Status:      "active",
		SetupStatus: models.SetupScaffolding,
	}
	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	// Create initial chat
	chat := &models.Chat{
		ProjectID:  project.ID,
		Title:      "Initial chat",
		BranchName: "main",
		Status:     "active",
		IsActive:   true,
	}
	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, fmt.Errorf("create initial chat: %w", err)
	}

	// Launch async setup
	go s.setupProject(project.ID)

	project.Chats = []models.Chat{*chat}
	return project, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	return s.projectRepo.FindByID(ctx, id)
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]*models.Project, error) {
	return s.projectRepo.FindAll(ctx)
}

func (s *ProjectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find project: %w", err)
	}

	// Stop and remove sandbox
	if project.Sandbox != nil && project.Sandbox.ContainerID != "" {
		s.sandboxMgr.Remove(ctx, project.Sandbox.ContainerID, project.Sandbox.Port)
		s.sandboxRepo.Delete(ctx, project.Sandbox.ID)
	}

	// Delete chats and messages
	for _, chat := range project.Chats {
		s.chatRepo.Delete(ctx, chat.ID)
	}

	return s.projectRepo.Delete(ctx, id)
}

func (s *ProjectService) setupProject(projectID uuid.UUID) {
	ctx := context.Background()
	projectDir := filepath.Join(s.dataDir, projectID.String())

	s.emitSetup(projectID, models.SetupScaffolding, "Scaffolding Next.js project...")

	// Scaffold
	if err := sandbox.ScaffoldNextApp(projectDir); err != nil {
		s.logger.Error("Scaffold failed", zap.Error(err))
		s.projectRepo.UpdateSetupStatus(ctx, projectID, models.SetupFailed, err.Error())
		s.emitSetup(projectID, models.SetupFailed, err.Error())
		return
	}

	s.emitSetup(projectID, models.SetupInstalling, "Installing dependencies...")

	// Git init
	if err := s.gitService.Init(projectDir); err != nil {
		s.logger.Error("Git init failed", zap.Error(err))
		s.projectRepo.UpdateSetupStatus(ctx, projectID, models.SetupFailed, err.Error())
		s.emitSetup(projectID, models.SetupFailed, err.Error())
		return
	}

	s.emitSetup(projectID, models.SetupStarting, "Starting dev server...")

	// Start sandbox
	containerID, port, err := s.sandboxMgr.CreateAndStart(ctx, projectID.String(), projectDir)
	if err != nil {
		s.logger.Error("Sandbox start failed", zap.Error(err))
		s.projectRepo.UpdateSetupStatus(ctx, projectID, models.SetupFailed, err.Error())
		s.emitSetup(projectID, models.SetupFailed, err.Error())
		return
	}

	// Save sandbox record
	sbx := &models.Sandbox{
		ProjectID:    projectID,
		ContainerID:  containerID,
		Port:         port,
		Status:       models.SandboxRunning,
		LastActiveAt: time.Now(),
	}
	if err := s.sandboxRepo.Create(ctx, sbx); err != nil {
		s.logger.Error("Save sandbox failed", zap.Error(err))
	}

	s.projectRepo.UpdateSetupStatus(ctx, projectID, models.SetupReady, "")
	s.emitSetup(projectID, models.SetupReady, "Project ready!")
	s.logger.Info("Project setup complete", zap.String("project_id", projectID.String()))
}

func (s *ProjectService) emitSetup(projectID uuid.UUID, status, message string) {
	if s.onSetupEvent != nil {
		s.onSetupEvent(projectID, status, message)
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric chars except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s := result.String()
	// Trim trailing hyphens
	s = strings.TrimRight(s, "-")
	if s == "" {
		s = uuid.New().String()[:8]
	}
	return s
}
