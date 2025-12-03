# Makefile for local development environment

.PHONY: help start stop logs restart rebuild scale test clean

help: ## Show this help message
	@echo "Winspire Local Development"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

start: ## Start all services
	@./start.sh

stop: ## Stop all services
	@./stop.sh

logs: ## Show logs from all services
	@docker-compose logs -f

logs-service: ## Show logs from a specific service (usage: make logs-service SERVICE=competition-host-stream)
	@docker-compose logs -f $(SERVICE)

restart: ## Restart all services
	@docker-compose restart

rebuild: ## Rebuild and restart all services
	@docker-compose up --build -d

scale: ## Start with scaled instances (2x competition-host-stream)
	@echo "🚀 Starting with scaled instances..."
	@docker-compose --profile scale up -d
	@echo "✅ Started 2x competition-host-stream instances"

scale-down: ## Stop scaled instances
	@docker-compose --profile scale down

test: ## Run basic health checks
	@echo "🔍 Testing services..."
	@echo ""
	@echo "Competition Host Stream:"
	@curl -s http://localhost/v1/cups || echo "❌ Failed"
	@echo ""
	@echo "Game Management:"
	@curl -s http://localhost/v1/games || echo "❌ Failed"
	@echo ""
	@echo "✅ Tests complete"

dashboard: ## Open Traefik dashboard
	@open http://localhost:8080 || xdg-open http://localhost:8080

redis-cli: ## Connect to Redis CLI
	@docker exec -it winspire-redis redis-cli

db-connect: ## Connect to database (Supabase)
	@psql postgresql://postgres:postgres@localhost:54322/postgres

clean: ## Stop and remove all containers, volumes, and images
	@echo "🧹 Cleaning up..."
	@docker-compose down -v --rmi all
	@echo "✅ Cleanup complete"

setup: ## Initial setup (copy env file)
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "✅ Created .env from env.example"; \
		echo "⚠️  Please edit .env with your Supabase credentials"; \
	else \
		echo "⚠️  .env already exists"; \
	fi


