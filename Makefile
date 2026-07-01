.PHONY: up down run build tidy seed-es lint ui

# Start all infrastructure (Postgres, Redis, ES, MinIO)
up:
	docker compose up -d
	@echo "Waiting for services..."
	@sleep 5

# Stop all infrastructure
down:
	docker compose down

# Run the API server locally
run:
	go run ./cmd/api/main.go

# Build binary
build:
	go build -o bin/api ./cmd/api/main.go

# Tidy go modules
tidy:
	go mod tidy

# Run database migrations
migrate:
	psql -U netflix -d mini_netflix -f migrations/sql/001_init.sql
	psql -U netflix -d mini_netflix -f migrations/sql/002_seed.sql

# Seed Elasticsearch
seed-es:
	bash scripts/es_seed.sh

# Full local dev setup
dev: up tidy seed-es run

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Open UI tools in browser
ui:
	@echo "Opening UI tools..."
	open http://localhost:5540   # RedisInsight
	open http://localhost:5601   # Kibana

# Show API routes summary
routes:
	@echo ""
	@echo "  Auth"
	@echo "    POST   /api/v1/auth/register"
	@echo "    POST   /api/v1/auth/login"
	@echo ""
	@echo "  Content"
	@echo "    POST   /api/v1/content              (auth)"
	@echo "    GET    /api/v1/content/:id"
	@echo "    POST   /api/v1/content/:id/watch"
	@echo "    POST   /api/v1/content/:id/rate     (auth)"
	@echo "    POST   /api/v1/content/:id/watchlist (auth)"
	@echo ""
	@echo "  Search"
	@echo "    GET    /api/v1/search?q=avengers"
	@echo "    GET    /api/v1/search/autocomplete?q=ave"
	@echo ""
	@echo "  Recommendations"
	@echo "    GET    /api/v1/recommendations/popular"
	@echo "    GET    /api/v1/recommendations/top-rated"
	@echo ""
	@echo "  User"
	@echo "    GET    /api/v1/user/watchlist        (auth)"
	@echo ""
