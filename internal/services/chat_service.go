package services

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/JuLima14/claude-engine/events"
	"github.com/JuLima14/claude-engine/runner"
	"github.com/JuLima14/pinfra-studio/internal/domain/models"
	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChatService struct {
	chatRepo    repositories.ChatRepository
	messageRepo repositories.MessageRepository
	projectRepo repositories.ProjectRepository
	sandboxRepo repositories.SandboxRepository
	gitService  *GitService
	broadcaster *events.Broadcaster
	logger      *zap.Logger
	dataDir     string
	runners     map[uuid.UUID]*runner.Runner // projectID -> runner (one per project)
}

func NewChatService(
	chatRepo repositories.ChatRepository,
	messageRepo repositories.MessageRepository,
	projectRepo repositories.ProjectRepository,
	sandboxRepo repositories.SandboxRepository,
	gitService *GitService,
	broadcaster *events.Broadcaster,
	logger *zap.Logger,
	dataDir string,
) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		projectRepo: projectRepo,
		sandboxRepo: sandboxRepo,
		gitService:  gitService,
		broadcaster: broadcaster,
		logger:      logger,
		dataDir:     dataDir,
		runners:     make(map[uuid.UUID]*runner.Runner),
	}
}

func (s *ChatService) CreateChat(ctx context.Context, projectID uuid.UUID) (*models.Chat, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("find project: %w", err)
	}

	count, _ := s.chatRepo.CountByProject(ctx, projectID)
	branchName := fmt.Sprintf("chat-%d", count+1)

	projectDir := filepath.Join(s.dataDir, project.ID.String())
	if err := s.gitService.CreateBranch(projectDir, branchName); err != nil {
		return nil, fmt.Errorf("create branch: %w", err)
	}

	chat := &models.Chat{
		ProjectID:  projectID,
		Title:      fmt.Sprintf("Chat %d", count+1),
		BranchName: branchName,
		Status:     "active",
		IsActive:   false,
	}
	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	// Activate this chat
	if err := s.chatRepo.SetActive(ctx, projectID, chat.ID); err != nil {
		return nil, fmt.Errorf("set active: %w", err)
	}
	chat.IsActive = true

	return chat, nil
}

func (s *ChatService) GetChat(ctx context.Context, chatID uuid.UUID) (*models.Chat, error) {
	return s.chatRepo.FindByID(ctx, chatID)
}

func (s *ChatService) ListChats(ctx context.Context, projectID uuid.UUID) ([]*models.Chat, error) {
	return s.chatRepo.FindByProject(ctx, projectID)
}

func (s *ChatService) ActivateChat(ctx context.Context, projectID, chatID uuid.UUID) error {
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}

	projectDir := filepath.Join(s.dataDir, projectID.String())

	// Commit current changes before switching
	s.gitService.CommitAll(projectDir, "Auto-save before branch switch")

	// Checkout the chat's branch
	if err := s.gitService.Checkout(projectDir, chat.BranchName); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	return s.chatRepo.SetActive(ctx, projectID, chatID)
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID uuid.UUID) error {
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}

	count, _ := s.chatRepo.CountByProject(ctx, chat.ProjectID)
	if count <= 1 {
		return fmt.Errorf("cannot delete the last chat")
	}

	// Delete git branch if not main
	if chat.BranchName != "main" {
		projectDir := filepath.Join(s.dataDir, chat.ProjectID.String())
		s.gitService.Checkout(projectDir, "main")
		s.gitService.DeleteBranch(projectDir, chat.BranchName)
	}

	return s.chatRepo.Delete(ctx, chatID)
}

func (s *ChatService) SendMessage(ctx context.Context, chatID uuid.UUID, content string) error {
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("find chat: %w", err)
	}

	project, err := s.projectRepo.FindByID(ctx, chat.ProjectID)
	if err != nil {
		return fmt.Errorf("find project: %w", err)
	}

	// Save user message
	userMsg := &models.Message{
		ChatID:  chatID,
		Role:    models.RoleUser,
		Content: content,
	}
	if err := s.messageRepo.Create(ctx, userMsg); err != nil {
		return fmt.Errorf("save user message: %w", err)
	}

	// Broadcast user message
	s.broadcaster.Broadcast(map[string]interface{}{
		"type":    "user_message",
		"content": content,
	}, chatID.String())

	// Get or create runner for this project
	r := s.getOrCreateRunner(project)

	// Build dynamic system prompt with sandbox context
	var systemPrompt string
	if sbx, err := s.sandboxRepo.FindByProject(ctx, project.ID); err == nil {
		s.sandboxRepo.TouchLastActive(ctx, sbx.ID)
		systemPrompt = fmt.Sprintf(`You are working inside Pinfra Studio, an AI-powered app builder.

IMPORTANT CONTEXT:
- This is a Next.js project with App Router (src/app/)
- A live dev server is running at http://localhost:%d with hot reload
- The user can see a live preview of every change you make instantly
- You are editing files directly on the filesystem — changes appear in the preview automatically
- Use TypeScript, Tailwind CSS, and the App Router conventions
- Keep components in src/app/ or src/components/
- Do NOT run 'npm run dev' — the dev server is already running
- Do NOT run 'npm install' unless adding a new dependency
- After making changes, briefly describe what you changed

CRITICAL RULES:
- Do NOT use any Skills (no Skill tool calls)
- Do NOT use brainstorming, planning, or design skills
- Do NOT ask to show mockups or visual companions in the browser
- Do NOT use Agent, EnterPlanMode, or any meta-tools
- Do NOT ask questions — just make the changes directly
- Be concise — the user sees the result live in the preview
- Just write code. No planning, no brainstorming, no asking for permission.`, sbx.Port)
	}

	// Run Claude async (pass session ID for conversation continuity)
	go s.runClaude(chatID, r, content, systemPrompt, chat.ClaudeSessionID)

	return nil
}

func (s *ChatService) CancelGeneration(projectID uuid.UUID) {
	if r, ok := s.runners[projectID]; ok {
		r.Cancel()
	}
}

func (s *ChatService) getOrCreateRunner(project *models.Project) *runner.Runner {
	if r, ok := s.runners[project.ID]; ok {
		return r
	}

	projectDir := filepath.Join(s.dataDir, project.ID.String())
	r := runner.NewRunner(runner.Config{
		WorkDir:  projectDir,
		Model:    "sonnet",
		MaxTurns: 50,
	})
	s.runners[project.ID] = r
	return r
}

func (s *ChatService) runClaude(chatID uuid.UUID, r *runner.Runner, prompt string, systemPrompt string, claudeSessionID string) {
	ctx := context.Background()
	channelID := chatID.String()
	var assistantText string

	result, err := r.Run(ctx, runner.RunOptions{
		Prompt:       prompt,
		SystemPrompt: systemPrompt,
		SessionID:    claudeSessionID,
		OnStream: func(chunk events.StreamChunk) {
			switch chunk.Type {
			case events.ChunkText:
				assistantText += chunk.Text
			case events.ChunkToolUse:
				// Save tool use as separate message
				inputJSON, _ := json.Marshal(chunk.ToolInput)
				inputStr := string(inputJSON)
				toolMsg := &models.Message{
					ChatID:    chatID,
					Role:      models.RoleTool,
					ToolName:  chunk.ToolName,
					ToolInput: &inputStr,
				}
				s.messageRepo.Create(ctx, toolMsg)
			}
			// Forward all chunks to SSE
			s.broadcaster.Broadcast(chunk, channelID)
		},
	})

	if err != nil {
		s.logger.Error("Claude run failed", zap.Error(err))
		s.broadcaster.Broadcast(events.StreamChunk{
			Type: events.ChunkError,
			Text: err.Error(),
		}, channelID)
		return
	}

	// Save assistant message
	if assistantText != "" {
		msg := &models.Message{
			ChatID:  chatID,
			Role:    models.RoleAssistant,
			Content: assistantText,
		}
		s.messageRepo.Create(ctx, msg)
	}

	// Update claude session ID on chat
	chat, _ := s.chatRepo.FindByID(ctx, chatID)
	if chat != nil && result.ClaudeSessionID != "" {
		chat.ClaudeSessionID = result.ClaudeSessionID
		chat.UpdatedAt = time.Now()
		s.chatRepo.Update(ctx, chat)
	}

	// Auto-commit after Claude finishes
	projectDir := filepath.Join(s.dataDir, chat.ProjectID.String())
	s.gitService.CommitAll(projectDir, fmt.Sprintf("Changes from chat: %s", prompt[:min(50, len(prompt))]))

	s.logger.Info("Claude run complete",
		zap.String("chat_id", chatID.String()),
		zap.Bool("success", result.Success),
		zap.Int64("duration_ms", result.DurationMs))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
