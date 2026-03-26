package services

import (
	"context"
	"fmt"
	"net/http"
	"os"
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

	// Step 1: Prepare project directory (fast — just creates dir + CLAUDE.md)
	s.emitSetup(projectID, models.SetupScaffolding, "Preparing project directory...")

	if err := sandbox.ScaffoldNextApp(projectDir); err != nil {
		s.logger.Error("Scaffold failed", zap.Error(err))
		s.projectRepo.UpdateSetupStatus(ctx, projectID, models.SetupFailed, err.Error())
		s.emitSetup(projectID, models.SetupFailed, err.Error())
		return
	}

	// Step 2: Start sandbox container (it will scaffold Next.js inside Linux)
	s.emitSetup(projectID, models.SetupScaffolding, "Starting container and scaffolding Next.js...")

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

	// Step 3: Wait for scaffold to complete inside container (poll for package.json)
	s.emitSetup(projectID, models.SetupInstalling, "Installing dependencies inside container...")

	packageJSON := filepath.Join(projectDir, "package.json")
	maxWait := 120 * time.Second
	pollInterval := 3 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		if _, err := os.Stat(packageJSON); err == nil {
			break
		}
		time.Sleep(pollInterval)
	}

	// Verify scaffold succeeded
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		s.logger.Error("Container scaffold timed out — no package.json after 2 min")
		s.projectRepo.UpdateSetupStatus(ctx, projectID, models.SetupFailed, "Scaffold timed out")
		s.emitSetup(projectID, models.SetupFailed, "Scaffold timed out after 2 minutes")
		return
	}

	// Step 4: Wait for dev server to be ready (poll HTTP)
	s.emitSetup(projectID, models.SetupStarting, "Waiting for dev server...")

	devURL := fmt.Sprintf("http://localhost:%d", port)
	deadline = time.Now().Add(90 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := http.Get(devURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				break
			}
		}
		time.Sleep(pollInterval)
	}

	// Step 5: Git init after scaffold is done
	if err := s.gitService.Init(projectDir); err != nil {
		s.logger.Warn("Git init failed (non-fatal)", zap.Error(err))
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
