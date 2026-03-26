package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	DataDir        string
	SandboxImage   string
	SandboxPortMin int
	SandboxPortMax int

	// Auth0 (shared with infra-platform)
	Auth0Domain string
}

func Load() *Config {
	dataDir := getEnv("DATA_DIR", "./data/projects")
	// Docker bind mounts require absolute paths
	if !filepath.IsAbs(dataDir) {
		abs, err := filepath.Abs(dataDir)
		if err == nil {
			dataDir = abs
		}
	}

	return &Config{
		Port:           getEnv("PORT", "8090"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/infraplatform?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		DataDir:        dataDir,
		SandboxImage:   getEnv("SANDBOX_IMAGE", "pinfra-sandbox:latest"),
		SandboxPortMin: 3100,
		SandboxPortMax: 3999,
		Auth0Domain:    getEnv("AUTH0_DOMAIN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
