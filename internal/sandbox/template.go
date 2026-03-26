package sandbox

import (
	"fmt"
	"os"
)

// ScaffoldNextApp prepares the project directory.
// The actual Next.js scaffold runs INSIDE the Docker container (sandbox-entrypoint.sh)
// to avoid native module architecture mismatches (macOS host vs Linux container).
// The dir must be EMPTY for create-next-app to work.
func ScaffoldNextApp(projectDir string) error {
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	// Dir must be empty — CLAUDE.md is written by entrypoint after scaffold
	return nil
}
