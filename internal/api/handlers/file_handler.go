package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/JuLima14/pinfra-studio/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FileEntry represents a node in the directory tree.
type FileEntry struct {
	Name     string       `json:"name"`
	Path     string       `json:"path"`
	IsDir    bool         `json:"isDir"`
	Children []*FileEntry `json:"children,omitempty"`
}

// excludedDirs contains directories to skip when building the file tree.
var excludedDirs = map[string]bool{
	"node_modules": true,
	".next":        true,
	".git":         true,
	".cache":       true,
	"dist":         true,
	"build":        true,
	".turbo":       true,
}

type FileHandler struct {
	projectRepo repositories.ProjectRepository
	dataDir     string
	logger      *zap.Logger
}

func NewFileHandler(
	projectRepo repositories.ProjectRepository,
	dataDir string,
	logger *zap.Logger,
) *FileHandler {
	return &FileHandler{
		projectRepo: projectRepo,
		dataDir:     dataDir,
		logger:      logger,
	}
}

// FileTree GET /api/v1/projects/:id/files
func (h *FileHandler) FileTree(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	projectDir := filepath.Join(h.dataDir, projectID.String())
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project directory not found"})
	}

	tree, err := buildFileTree(projectDir, projectDir)
	if err != nil {
		h.logger.Error("build file tree failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tree)
}

// FileContent GET /api/v1/projects/:id/files/*path
func (h *FileHandler) FileContent(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	relPath := c.Params("*")
	if relPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path is required"})
	}

	// Sanitize path: prevent directory traversal
	projectDir := filepath.Join(h.dataDir, projectID.String())
	absPath := filepath.Join(projectDir, filepath.Clean("/"+relPath))

	// Ensure absPath is within projectDir
	if !strings.HasPrefix(absPath, projectDir) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if info.IsDir() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path is a directory"})
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		h.logger.Error("read file failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"content": string(content), "path": relPath})
}

// buildFileTree recursively builds a file tree rooted at dir.
// rootDir is used to compute relative paths for the Path field.
func buildFileTree(dir, rootDir string) ([]*FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []*FileEntry
	for _, entry := range entries {
		name := entry.Name()

		// Skip excluded directories
		if entry.IsDir() && excludedDirs[name] {
			continue
		}
		// Skip hidden files/dirs (e.g. .git when not in excludedDirs)
		if strings.HasPrefix(name, ".") && !entry.IsDir() {
			continue
		}

		absPath := filepath.Join(dir, name)
		relPath, _ := filepath.Rel(rootDir, absPath)

		node := &FileEntry{
			Name:  name,
			Path:  relPath,
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			children, err := buildFileTree(absPath, rootDir)
			if err == nil {
				node.Children = children
			}
		}

		result = append(result, node)
	}
	return result, nil
}
