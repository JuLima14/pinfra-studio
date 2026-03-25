package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

type Manager struct {
	ports  *PortAllocator
	image  string
	logger *zap.Logger
}

func NewManager(image string, portMin, portMax int, logger *zap.Logger) *Manager {
	return &Manager{
		ports:  NewPortAllocator(portMin, portMax),
		image:  image,
		logger: logger,
	}
}

func (m *Manager) CreateAndStart(ctx context.Context, projectID string, hostDir string) (containerID string, port int, err error) {
	port, err = m.ports.Allocate()
	if err != nil {
		return "", 0, fmt.Errorf("allocate port: %w", err)
	}

	containerName := fmt.Sprintf("pinfra-sandbox-%s", projectID[:8])

	// Remove existing container if any
	exec.CommandContext(ctx, "docker", "rm", "-f", containerName).Run()

	args := []string{
		"run", "-d",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/app", hostDir),
		"-p", fmt.Sprintf("%d:3000", port),
		m.image,
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		m.ports.Release(port)
		return "", 0, fmt.Errorf("docker run: %s: %w", string(out), err)
	}

	cid := strings.TrimSpace(string(out))
	if len(cid) > 12 {
		cid = cid[:12]
	}

	m.logger.Info("Sandbox started",
		zap.String("container_id", cid),
		zap.String("project_id", projectID),
		zap.Int("port", port))

	return cid, port, nil
}

func (m *Manager) Stop(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "stop", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker stop: %s: %w", string(out), err)
	}
	return nil
}

func (m *Manager) Start(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "start", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker start: %s: %w", string(out), err)
	}
	return nil
}

func (m *Manager) Remove(ctx context.Context, containerID string, port int) error {
	exec.CommandContext(ctx, "docker", "stop", containerID).Run()
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker rm: %s: %w", string(out), err)
	}
	m.ports.Release(port)
	return nil
}

func (m *Manager) Logs(ctx context.Context, containerID string, lines int) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "logs", "--tail", fmt.Sprintf("%d", lines), containerID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker logs: %s: %w", string(out), err)
	}
	return string(out), nil
}

func (m *Manager) IsRunning(ctx context.Context, containerID string) bool {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", containerID)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}
