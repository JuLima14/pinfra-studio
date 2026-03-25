package config

import "os"

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	DataDir        string
	SandboxImage   string
	SandboxPortMin int
	SandboxPortMax int
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8090"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/pinfra_studio?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6380/0"),
		DataDir:        getEnv("DATA_DIR", "./data/projects"),
		SandboxImage:   getEnv("SANDBOX_IMAGE", "pinfra-sandbox:latest"),
		SandboxPortMin: 3100,
		SandboxPortMax: 3999,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
