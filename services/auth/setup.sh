#!/bin/bash

# Setup script for Auth Service
# This script installs dependencies and sets up the project

set -e

echo "=== Setting up Auth Service ==="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.25.4 or later."
    exit 1
fi

echo "✓ Go is installed: $(go version)"
echo ""

# Install dependencies
echo "Installing Go dependencies..."
cd "$(dirname "$0")"

# Install dependencies for auth service
echo "  → Installing auth service dependencies..."
go get github.com/gin-gonic/gin
go get github.com/kelseyhightower/envconfig
go get github.com/jackc/pgx/v5
go get github.com/supabase-community/supabase-go
go get github.com/golang-jwt/jwt/v5

# Install dependencies for auth library
echo "  → Installing auth library dependencies..."
cd ../libs/go/auth
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5

# Run go mod tidy
cd ../../services/auth
echo "  → Running go mod tidy..."
go mod tidy

cd ../libs/go/auth
go mod tidy

echo ""
echo "✓ Dependencies installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Create .env file: cp .env.example .env"
echo "  2. Add Supabase credentials to .env"
echo "  3. Start Supabase: cd ../../supabase && docker-compose up -d"
echo "  4. Run the service: make run"
echo ""

