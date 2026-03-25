package services

import (
	"context"
	"fmt"
	"time"

	"github.com/JuLima14/pinfra-studio/internal/domain/models"
	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/JuLima14/pinfra-studio/internal/sandbox"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SandboxService struct {
	sandboxRepo repositories.SandboxRepository
	sandboxMgr  *sandbox.Manager
	logger      *zap.Logger
}

func NewSandboxService(
	sandboxRepo repositories.SandboxRepository,
	sandboxMgr *sandbox.Manager,
	logger *zap.Logger,
) *SandboxService {
	return &SandboxService{
		sandboxRepo: sandboxRepo,
		sandboxMgr:  sandboxMgr,
		logger:      logger,
	}
}

func (s *SandboxService) GetStatus(ctx context.Context, projectID uuid.UUID) (*models.Sandbox, error) {
	sbx, err := s.sandboxRepo.FindByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("find sandbox: %w", err)
	}

	// Verify actual Docker state matches DB
	if sbx.ContainerID != "" {
		isRunning := s.sandboxMgr.IsRunning(ctx, sbx.ContainerID)
		if isRunning && sbx.Status != models.SandboxRunning {
			sbx.Status = models.SandboxRunning
			s.sandboxRepo.UpdateStatus(ctx, sbx.ID, models.SandboxRunning)
		} else if !isRunning && sbx.Status == models.SandboxRunning {
			sbx.Status = models.SandboxStopped
			s.sandboxRepo.UpdateStatus(ctx, sbx.ID, models.SandboxStopped)
		}
	}

	return sbx, nil
}

func (s *SandboxService) Start(ctx context.Context, projectID uuid.UUID) error {
	sbx, err := s.sandboxRepo.FindByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("find sandbox: %w", err)
	}
	if sbx.Status == models.SandboxRunning {
		return nil // already running
	}
	if err := s.sandboxMgr.Start(ctx, sbx.ContainerID); err != nil {
		return fmt.Errorf("start sandbox: %w", err)
	}
	s.sandboxRepo.UpdateStatus(ctx, sbx.ID, models.SandboxRunning)
	s.sandboxRepo.TouchLastActive(ctx, sbx.ID)
	return nil
}

func (s *SandboxService) Stop(ctx context.Context, projectID uuid.UUID) error {
	sbx, err := s.sandboxRepo.FindByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("find sandbox: %w", err)
	}
	if err := s.sandboxMgr.Stop(ctx, sbx.ContainerID); err != nil {
		return fmt.Errorf("stop sandbox: %w", err)
	}
	s.sandboxRepo.UpdateStatus(ctx, sbx.ID, models.SandboxStopped)
	return nil
}

func (s *SandboxService) GetURL(ctx context.Context, projectID uuid.UUID) (string, error) {
	sbx, err := s.sandboxRepo.FindByProject(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("find sandbox: %w", err)
	}
	return fmt.Sprintf("http://localhost:%d", sbx.Port), nil
}

func (s *SandboxService) GetLogs(ctx context.Context, projectID uuid.UUID, lines int) (string, error) {
	sbx, err := s.sandboxRepo.FindByProject(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("find sandbox: %w", err)
	}
	return s.sandboxMgr.Logs(ctx, sbx.ContainerID, lines)
}

func (s *SandboxService) CleanupStale(ctx context.Context) error {
	stale, err := s.sandboxRepo.FindStale(ctx, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("find stale: %w", err)
	}
	for _, sbx := range stale {
		s.logger.Info("Stopping stale sandbox",
			zap.String("sandbox_id", sbx.ID.String()),
			zap.String("project_id", sbx.ProjectID.String()))
		s.sandboxMgr.Stop(ctx, sbx.ContainerID)
		s.sandboxRepo.UpdateStatus(ctx, sbx.ID, models.SandboxStopped)
	}
	return nil
}
