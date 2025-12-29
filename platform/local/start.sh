#!/bin/bash
# Quick start script for local development
# Uses LocalStack to emulate AWS ALB for path-based routing

set -e

echo "Starting Winspire Local Development Environment"
echo "Using LocalStack ALB for routing"
echo ""

# Check if we're in the right directory
if [ ! -f "docker-compose.yaml" ]; then
    echo "Error: Must be run from platform/local/ directory"
    exit 1
fi

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "Warning: .env file not found"
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo "Created .env - Please edit it with your Supabase credentials"
    echo ""
    echo "   Get credentials from:"
    echo "   1. Run: cd ../supabase && supabase start"
    echo "   2. Copy JWT secret and anon key to .env"
    echo ""
    read -p "Press Enter when ready to continue..."
fi

# Check if Supabase is running
echo "Checking Supabase status..."
if ! nc -z localhost 54322 2>/dev/null; then
    echo "Warning: Supabase doesn't appear to be running"
    echo "   Run: cd ../supabase && supabase start"
    read -p "Press Enter to continue anyway, or Ctrl+C to exit..."
else
    echo "Supabase is running"
fi

echo ""
echo "Starting Docker services..."
docker-compose up -d

echo ""
echo "Waiting for LocalStack to be healthy..."
max_attempts=30
attempt=0
while ! curl -sf http://localhost:4566/_localstack/health > /dev/null 2>&1; do
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
        echo "Error: LocalStack failed to start after ${max_attempts} attempts"
        echo "Check logs: docker-compose logs localstack"
        exit 1
    fi
    echo "  Waiting for LocalStack... (attempt $attempt/$max_attempts)"
    sleep 2
done
echo "LocalStack is healthy"

echo ""
echo "Waiting for services to be ready..."
sleep 10

echo ""
echo "============================================"
echo "Services started!"
echo "============================================"
echo ""
echo "LocalStack Health: http://localhost:4566/_localstack/health"
echo "API Gateway (ALB): http://localhost"
echo ""
echo "Routing:"
echo "  /v1/games*, /v1/g/*, /v1/admin* -> game-management:8085"
echo "  /v1/matchmaking*                -> matchmaking:8081"
echo "  /v1/* (default)                 -> tournament:8089"
echo ""
echo "Test endpoints:"
echo "  curl http://localhost/v1/cups              # tournament"
echo "  curl http://localhost/v1/games             # game-management"
echo "  curl http://localhost/v1/matchmaking/health # matchmaking"
echo ""
echo "Direct service access (for E2E tests):"
echo "  http://localhost:8089 -> tournament"
echo "  http://localhost:8088 -> matchmaking"
echo ""
echo "Useful commands:"
echo "  make logs           # View all logs"
echo "  make alb-status     # Show ALB status"
echo "  make alb-targets    # Show registered targets"
echo "  make health         # Check service health"
echo "  make test           # Run health checks"
echo "  docker-compose down # Stop all services"
echo ""
