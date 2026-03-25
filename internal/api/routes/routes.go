package routes

import (
	"github.com/JuLima14/pinfra-studio/internal/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// Register mounts all API routes on the Fiber app.
func Register(
	app *fiber.App,
	projectHandler *handlers.ProjectHandler,
	chatHandler *handlers.ChatHandler,
	messageHandler *handlers.MessageHandler,
	sandboxHandler *handlers.SandboxHandler,
	fileHandler *handlers.FileHandler,
	githubHandler *handlers.GitHubHandler,
) {
	api := app.Group("/api/v1")

	// Health
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Projects
	projects := api.Group("/projects")
	projects.Get("/", projectHandler.ListProjects)
	projects.Post("/", projectHandler.CreateProject)
	projects.Get("/:id", projectHandler.GetProject)
	projects.Delete("/:id", projectHandler.DeleteProject)

	// Chats (nested under project)
	chats := projects.Group("/:id/chats")
	chats.Get("/", chatHandler.ListChats)
	chats.Post("/", chatHandler.CreateChat)
	chats.Get("/:chatId", chatHandler.GetChat)
	chats.Delete("/:chatId", chatHandler.DeleteChat)
	chats.Post("/:chatId/activate", chatHandler.ActivateChat)

	// SSE Events (nested under chats)
	chats.Get("/:chatId/events", messageHandler.StreamEvents)

	// Messages (nested under chats)
	messages := chats.Group("/:chatId/messages")
	messages.Post("/", messageHandler.SendMessage)
	messages.Post("/cancel", messageHandler.CancelMessage)

	// Sandbox
	sandbox := projects.Group("/:id/sandbox")
	sandbox.Get("/status", sandboxHandler.GetStatus)
	sandbox.Post("/start", sandboxHandler.StartSandbox)
	sandbox.Post("/stop", sandboxHandler.StopSandbox)
	sandbox.Get("/url", sandboxHandler.GetURL)
	sandbox.Get("/logs", sandboxHandler.GetLogs)

	// Files
	files := projects.Group("/:id/files")
	files.Get("/", fileHandler.FileTree)
	files.Get("/*", fileHandler.FileContent)

	// GitHub
	github := projects.Group("/:id/github")
	github.Post("/connect", githubHandler.ConnectGitHub)
	github.Post("/push", githubHandler.PushToGitHub)
	github.Post("/pr", githubHandler.CreatePR)
}
