.PHONY: dev web-dev test docker-infra docker-infra-down sandbox-build clean help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run server with hot reload
	go run cmd/server/main.go

web-dev: ## Run frontend dev server
	cd web && npm run dev

test: ## Run all Go tests
	go test ./... -v -count=1

docker-infra: ## Start Postgres + Redis
	docker-compose -f docker/docker-compose.yml up -d

docker-infra-down: ## Stop Postgres + Redis
	docker-compose -f docker/docker-compose.yml down

sandbox-build: ## Build sandbox Docker image
	docker build -t pinfra-sandbox:latest -f docker/Dockerfile.sandbox docker/

clean: ## Clean build artifacts
	rm -rf bin/ data/ web/dist/
