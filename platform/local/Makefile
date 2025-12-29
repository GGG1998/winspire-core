# Makefile for local development environment
# Uses LocalStack to emulate AWS ALB for path-based routing

.PHONY: help start stop logs restart rebuild scale test clean alb-status alb-targets health

help: ## Show this help message
	@echo "Winspire Local Development (LocalStack ALB)"
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

logs-service: ## Show logs from a specific service (usage: make logs-service SERVICE=tournament)
	@docker-compose logs -f $(SERVICE)

logs-localstack: ## Show LocalStack logs
	@docker-compose logs -f localstack

logs-registrar: ## Show ALB registrar logs
	@docker-compose logs -f alb-registrar

restart: ## Restart all services
	@docker-compose restart

rebuild: ## Rebuild and restart all services
	@docker-compose up --build -d

scale: ## Start with scaled instances (2x tournament)
	@echo "Starting with scaled instances..."
	@docker-compose --profile scale up -d
	@echo "Started 2x tournament instances"

scale-down: ## Stop scaled instances
	@docker-compose --profile scale down

test: ## Run basic health checks via ALB routing
	@echo "Testing services via LocalStack ALB..."
	@echo ""
	@echo "Tournament (/v1/cups):"
	@curl -s http://localhost/v1/cups 2>/dev/null || echo "Failed (service may not be ready)"
	@echo ""
	@echo "Game Management (/v1/games):"
	@curl -s http://localhost/v1/games 2>/dev/null || echo "Failed (service may not be ready)"
	@echo ""
	@echo "Matchmaking (/v1/matchmaking/health):"
	@curl -s http://localhost/v1/matchmaking/health 2>/dev/null || echo "Failed (service may not be ready)"
	@echo ""
	@echo "Tests complete"

health: ## Check LocalStack and service health
	@echo "LocalStack Health:"
	@curl -s http://localhost:4566/_localstack/health | jq . 2>/dev/null || curl -s http://localhost:4566/_localstack/health
	@echo ""
	@echo "Service Health (direct):"
	@echo "  Tournament:      $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8089/healthz 2>/dev/null || echo 'down')"
	@echo "  Matchmaking:     $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8088/health 2>/dev/null || echo 'down')"

alb-status: ## Show LocalStack ALB status
	@echo "LocalStack ALB Status"
	@echo "====================="
	@echo ""
	@echo "Load Balancers:"
	@docker exec winspire-localstack awslocal elbv2 describe-load-balancers --query 'LoadBalancers[*].[LoadBalancerName,State.Code,DNSName]' --output table 2>/dev/null || echo "LocalStack not running"
	@echo ""
	@echo "Target Groups:"
	@docker exec winspire-localstack awslocal elbv2 describe-target-groups --query 'TargetGroups[*].[TargetGroupName,Port,HealthCheckPath]' --output table 2>/dev/null || echo "LocalStack not running"

alb-targets: ## Show registered ALB targets
	@echo "Registered ALB Targets"
	@echo "======================"
	@echo ""
	@for tg in tournament-tg game-management-tg matchmaking-tg; do \
		echo "$$tg:"; \
		docker exec winspire-localstack awslocal elbv2 describe-target-health \
			--target-group-arn $$(docker exec winspire-localstack awslocal elbv2 describe-target-groups --names $$tg --query 'TargetGroups[0].TargetGroupArn' --output text 2>/dev/null) \
			--query 'TargetHealthDescriptions[*].[Target.Id,Target.Port,TargetHealth.State]' --output table 2>/dev/null || echo "  Not available"; \
		echo ""; \
	done

alb-rules: ## Show ALB listener rules
	@echo "ALB Listener Rules"
	@echo "=================="
	@LISTENER=$$(docker exec winspire-localstack awslocal elbv2 describe-listeners \
		--load-balancer-arn $$(docker exec winspire-localstack awslocal elbv2 describe-load-balancers --names winspire-local-alb --query 'LoadBalancers[0].LoadBalancerArn' --output text 2>/dev/null) \
		--query 'Listeners[0].ListenerArn' --output text 2>/dev/null); \
	docker exec winspire-localstack awslocal elbv2 describe-rules --listener-arn $$LISTENER \
		--query 'Rules[*].[Priority,Conditions[0].PathPatternConfig.Values[0],Actions[0].TargetGroupArn]' --output table 2>/dev/null || echo "LocalStack not running"

redis-cli: ## Connect to Redis CLI
	@docker exec -it winspire-redis redis-cli

db-connect: ## Connect to database (Supabase)
	@psql postgresql://postgres:postgres@localhost:54322/postgres

clean: ## Stop and remove all containers, volumes, and images
	@echo "Cleaning up..."
	@docker-compose down -v --rmi all
	@echo "Cleanup complete"

setup: ## Initial setup (copy env file)
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "Created .env from env.example"; \
		echo "Please edit .env with your Supabase credentials"; \
	else \
		echo ".env already exists"; \
	fi

# Deployment targets (require AWS credentials)
deploy-dev: ## Deploy all services to dev environment
	@../../scripts/deploy.sh -e dev

deploy-staging: ## Deploy all services to staging environment
	@../../scripts/deploy.sh -e staging

deploy-prod: ## Deploy all services to production (with confirmation)
	@echo "WARNING: This will deploy to PRODUCTION!"
	@read -p "Are you sure? (yes/no): " confirm && [ "$$confirm" = "yes" ] && ../../scripts/deploy.sh -e prod -u || echo "Cancelled"

deploy-dry-run: ## Preview deployment without executing
	@../../scripts/deploy.sh -e dev --dry-run
