# Pinfra Studio

AI-powered app builder — generate Next.js apps through chat with live preview.

## Stack
- Go + Fiber v2 (HTTP framework)
- GORM + PostgreSQL (persistence)
- Redis (SSE buffering, pub/sub)
- Docker API (sandbox containers)
- claude-engine lib (Claude CLI orchestration)

## Architecture
- Server runs on HOST (not in Docker)
- Claude CLI runs as subprocess on host
- Sandbox containers run Next.js dev servers
- Project files on host, bind-mounted into containers

## Conventions
- Follow infra-platform patterns
- UUID primary keys
- Errors wrapped with context: fmt.Errorf("context: %w", err)
- Structured logging with zap
- Handlers: validate → service call → respond
- Services: validate → repo call → business logic → persist

## Commands
make dev          # Run server with hot reload
make web-dev      # Run frontend dev server
make test         # Run all tests
make docker-infra # Start Postgres + Redis
make sandbox-build # Build sandbox Docker image

## Ports
- API: 8090
- Frontend: 5173 (Vite)
- PostgreSQL: 5433
- Redis: 6380
- Sandboxes: 3100-3999
