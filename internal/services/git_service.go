package services

import (
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

type GitService struct {
	logger *zap.Logger
}

func NewGitService(logger *zap.Logger) *GitService {
	return &GitService{logger: logger}
}

// Init initializes a git repository in dir with an initial commit.
func (s *GitService) Init(dir string) error {
	if err := s.run(dir, "init"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	if err := s.run(dir, "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	if err := s.run(dir, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("git initial commit: %w", err)
	}
	return nil
}

// CreateBranch creates and switches to a new branch.
func (s *GitService) CreateBranch(dir, name string) error {
	if err := s.run(dir, "checkout", "-b", name); err != nil {
		return fmt.Errorf("git checkout -b %s: %w", name, err)
	}
	return nil
}

// Checkout stashes any uncommitted changes and switches to the given branch.
func (s *GitService) Checkout(dir, name string) error {
	// Stash changes so checkout doesn't fail on dirty working tree.
	// Ignore error: stash fails harmlessly when there's nothing to stash.
	s.run(dir, "stash") //nolint:errcheck

	if err := s.run(dir, "checkout", name); err != nil {
		return fmt.Errorf("git checkout %s: %w", name, err)
	}
	return nil
}

// CommitAll stages all changes and creates a commit with the given message.
// Returns without error if there is nothing to commit.
func (s *GitService) CommitAll(dir, message string) error {
	if err := s.run(dir, "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are staged changes; skip commit if not.
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = dir
	if err := cmd.Run(); err == nil {
		// Exit code 0 means no staged changes — nothing to commit.
		return nil
	}

	if err := s.run(dir, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// DeleteBranch forcefully deletes a local branch.
func (s *GitService) DeleteBranch(dir, name string) error {
	if err := s.run(dir, "branch", "-D", name); err != nil {
		return fmt.Errorf("git branch -D %s: %w", name, err)
	}
	return nil
}

// CurrentBranch returns the name of the currently checked-out branch.
func (s *GitService) CurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		s.logger.Error("git rev-parse failed",
			zap.String("dir", dir),
			zap.Error(err))
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// run executes a git command in the given directory, logging on error.
func (s *GitService) run(dir string, args ...string) error {
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
