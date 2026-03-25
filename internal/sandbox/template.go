package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ScaffoldNextApp(projectDir string) error {
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	cmd := exec.Command("npx", "create-next-app@latest", ".",
		"--typescript", "--tailwind", "--eslint", "--app",
		"--src-dir", "--import-alias", "@/*", "--use-npm",
	)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create-next-app: %w", err)
	}

	// Add shadcn/ui
	cmd = exec.Command("npx", "shadcn@latest", "init", "-d")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shadcn init: %w", err)
	}

	// Write CLAUDE.md for the generated project
	claudeMD := `# Project Instructions

You are building a Next.js application with the App Router.

## Stack
- Next.js 14+ with App Router (src/ directory)
- TypeScript (strict mode)
- Tailwind CSS for styling
- shadcn/ui component library (already installed)
- npm as package manager

## Rules
- Always use TypeScript
- Use the App Router (src/app/) not Pages Router
- Use shadcn/ui components from @/components/ui/
- Use Tailwind for styling, never inline styles or CSS modules
- Use server components by default, 'use client' only when needed
- Keep components small and focused
`
	if err := os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), []byte(claudeMD), 0644); err != nil {
		return fmt.Errorf("write CLAUDE.md: %w", err)
	}
	return nil
}
