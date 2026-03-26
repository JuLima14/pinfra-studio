package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"

	"github.com/JuLima14/claude-engine/events"
	"github.com/JuLima14/pinfra-studio/internal/api/handlers"
	"github.com/JuLima14/pinfra-studio/internal/api/middleware"
	"github.com/JuLima14/pinfra-studio/internal/api/routes"
	"github.com/JuLima14/pinfra-studio/internal/auth"
	"github.com/JuLima14/pinfra-studio/internal/config"
	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/JuLima14/pinfra-studio/internal/sandbox"
	"github.com/JuLima14/pinfra-studio/internal/services"
	"github.com/JuLima14/pinfra-studio/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 1b. Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create data dir: %v\n", err)
		os.Exit(1)
	}

	// 2. Connect PostgreSQL, auto-migrate
	db, err := storage.Connect(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	if err := storage.AutoMigrate(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to auto-migrate: %v\n", err)
		os.Exit(1)
	}

	// 3. Create zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 4. Create SSE broadcaster
	broadcaster := events.NewBroadcaster()

	// 5. Create sandbox Manager
	sandboxMgr := sandbox.NewManager(cfg.SandboxImage, cfg.SandboxPortMin, cfg.SandboxPortMax, logger)

	// 6. Create all repositories
	projectRepo := repositories.NewProjectRepository(db)
	chatRepo := repositories.NewChatRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	sandboxRepo := repositories.NewSandboxRepository(db)

	// 7. Create all services
	gitService := services.NewGitService(logger)
	githubService := services.NewGitHubService(logger)
	sandboxService := services.NewSandboxService(sandboxRepo, sandboxMgr, logger)
	chatService := services.NewChatService(
		chatRepo,
		messageRepo,
		projectRepo,
		sandboxRepo,
		gitService,
		broadcaster,
		logger,
		cfg.DataDir,
	)

	onSetupEvent := func(projectID uuid.UUID, status, message string) {
		broadcaster.Broadcast(map[string]interface{}{
			"type":       "setup_event",
			"project_id": projectID.String(),
			"status":     status,
			"message":    message,
		}, projectID.String())
	}

	projectService := services.NewProjectService(
		projectRepo,
		chatRepo,
		sandboxRepo,
		sandboxMgr,
		gitService,
		logger,
		cfg.DataDir,
		onSetupEvent,
	)

	// 8. Create handlers
	projectHandler := handlers.NewProjectHandler(projectService, logger)
	chatHandler := handlers.NewChatHandler(chatService, logger)
	messageHandler := handlers.NewMessageHandler(chatService, projectRepo, broadcaster, logger)
	sandboxHandler := handlers.NewSandboxHandler(sandboxService, projectRepo, logger)
	fileHandler := handlers.NewFileHandler(projectRepo, cfg.DataDir, logger)
	githubHandler := handlers.NewGitHubHandler(githubService, gitService, projectRepo, cfg.DataDir, logger)

	// 9. Setup JWT validator for Auth0
	var jwtValidator *auth.JWTValidator
	if cfg.Auth0Domain != "" {
		jwtValidator = auth.NewJWTValidator(cfg.Auth0Domain)
		logger.Info("Auth0 JWT validation enabled", zap.String("domain", cfg.Auth0Domain))
	} else {
		logger.Warn("AUTH0_DOMAIN not set — auth middleware disabled (dev mode)")
	}

	// 10. Setup Fiber app and routes
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Error("unhandled error", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(middleware.CORS())

	// Apply auth middleware if Auth0 is configured
	if jwtValidator != nil {
		app.Use("/api", middleware.AuthMiddleware(jwtValidator, db, logger))
	}

	routes.Register(app, projectHandler, chatHandler, messageHandler, sandboxHandler, fileHandler, githubHandler)

	// 10. Start cleanup goroutine (every 5 minutes)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctx := context.Background()
			if err := sandboxService.CleanupStale(ctx); err != nil {
				logger.Warn("cleanup stale sandboxes failed", zap.Error(err))
			}
		}
	}()

	// 11. Start Fiber with graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + cfg.Port
		logger.Info("server starting", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			logger.Error("server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down server")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
	logger.Info("server stopped")
}
