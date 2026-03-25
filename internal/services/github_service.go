package services

import (
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

type GitHubService struct {
	logger *zap.Logger
}

func NewGitHubService(logger *zap.Logger) *GitHubService {
	return &GitHubService{logger: logger}
}

// ConnectRepo ensures a remote named "origin" points to repoURL.
// If the remote already exists it updates its URL; otherwise it adds it.
func (s *GitHubService) ConnectRepo(dir, repoURL string) error {
	// Check if the remote already exists.
	checkCmd := exec.Command("git", "remote", "get-url", "origin")
	checkCmd.Dir = dir
	if err := checkCmd.Run(); err == nil {
		// Remote exists — update the URL.
		if err := s.run(dir, "remote", "set-url", "origin", repoURL); err != nil {
			return fmt.Errorf("git remote set-url: %w", err)
		}
	} else {
		// Remote does not exist — add it.
		if err := s.run(dir, "remote", "add", "origin", repoURL); err != nil {
			return fmt.Errorf("git remote add: %w", err)
		}
	}
	return nil
}

// Push pushes branchName to origin, setting the upstream tracking reference.
func (s *GitHubService) Push(dir, branchName string) error {
	if err := s.run(dir, "push", "-u", "origin", branchName); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

// CreatePR creates a pull request via the GitHub CLI and returns its output.
func (s *GitHubService) CreatePR(dir, title, body string) (string, error) {
	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
	)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("gh pr create failed",
			zap.String("dir", dir),
			zap.String("title", title),
			zap.String("output", string(out)),
			zap.Error(err))
		return "", fmt.Errorf("gh pr create: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// run executes a git command in the given directory, logging on error.
func (s *GitHubService) run(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("git command failed",
			zap.String("dir", dir),
			zap.Strings("args", args),
			zap.String("output", string(out)),
			zap.Error(err))
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
