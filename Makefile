.PHONY: help build up down logs migrate-up migrate-down restart clean test docs

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build Docker images
	docker compose build

up: ## Start all services
	docker compose up -d
	@echo "Services starting..."
	@echo "API will be available at http://localhost:8080"
	@echo "MinIO Console at http://localhost:9001 (admin/minioadmin)"
	@echo "PostgreSQL at localhost:5432"

down: ## Stop all services
	docker compose down

logs: ## View logs from all services
	docker compose logs -f

logs-api: ## View API logs
	docker compose logs -f api

logs-db: ## View database logs
	docker compose logs -f postgres

migrate-up: ## Run database migrations up
	docker compose exec api ./migrate -path=./migrations -database="${DATABASE_URL}" up

migrate-down: ## Rollback last migration
	docker compose exec api ./migrate -path=./migrations -database="${DATABASE_URL}" down 1

restart: ## Restart all services
	docker compose restart

restart-api: ## Restart API service
	docker compose restart api

clean: ## Stop and remove all containers, volumes, and images
	docker compose down -v
	docker compose rm -f

test: ## Run tests
	go test -v ./...

dev: ## Run in development mode without Docker
	go run cmd/api/main.go

create-bucket: ## Create MinIO bucket for development
	@echo "Creating blog-images bucket in MinIO..."
	@docker compose exec -T minio mc alias set local http://localhost:9000 minioadmin minioadmin 2>/dev/null || true
	@docker compose exec -T minio mc mb local/blog-images 2>/dev/null || echo "Bucket already exists"
	@docker compose exec -T minio mc anonymous set public local/blog-images
	@echo "Bucket created and set to public access"

docs: ## Regenerate Swagger docs (requires: go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/api/main.go -o docs

setup: up ## Complete setup - start services and initialize
	@echo "Waiting for services to be ready..."
	@sleep 10
	@$(MAKE) create-bucket
	@echo "Setup complete! API is ready at http://localhost:8080"
