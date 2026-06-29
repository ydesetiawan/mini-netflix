.PHONY: up down run build tidy seed-es lint

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

# Seed Elasticsearch
seed-es:
	bash scripts/es_seed.sh

# Full local dev setup
dev: up tidy seed-es run

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

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
