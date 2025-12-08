new-service-interactive:
	@echo "🚀 Creating new microservice (interactive)..."
	@echo ""
	@read -p "Service name (e.g. user-management): " name; \
	read -p "Service port (e.g. 8089): " port; \
	read -p "Database name (e.g. user_management): " db; \
	read -p "Use Redis? (true/false) [true]: " redis; redis=$${redis:-true}; \
	read -p "Use SQLC? (true/false) [true]: " sqlc; sqlc=$${sqlc:-true}; \
	cd services && cookiecutter template/ \
		--no-input \
		service_name="$$name" \
		service_port="$$port" \
		db_name="$$db" \
		use_redis="$$redis" \
		use_sqlc="$$sqlc" \
		go_module="github.com/winspire/$$name"
	@echo ""
	@echo "✅ Service created! Run 'make sync' to update go.work"

# ============================================================================
# Go Workspace Management
# ============================================================================

# Sync go.work with all services
sync:
	@echo "🔄 Syncing go.work..."
	@go work sync
	@echo "✅ go.work synced"

# Add new service to go.work
add-to-workspace:
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Usage: make add-to-workspace SERVICE=<service-name>"; \
		exit 1; \
	fi
	@go work use ./services/$(SERVICE)
	@echo "✅ Added ./services/$(SERVICE) to go.work"


dev-mini-admin:
	cd services/game-management && make run &
	cd frontends/mini-admin && yarn dev