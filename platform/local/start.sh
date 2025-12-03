#!/bin/bash
# Quick start script for local development

set -e

echo "🚀 Starting Winspire Local Development Environment"
echo ""

# Check if we're in the right directory
if [ ! -f "docker-compose.yaml" ]; then
    echo "❌ Error: Must be run from platform/local/ directory"
    exit 1
fi

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "⚠️  Warning: .env file not found"
    echo "📝 Creating .env from .env.example..."
    cp .env.example .env
    echo "✅ Created .env - Please edit it with your Supabase credentials"
    echo ""
    echo "   Get credentials from:"
    echo "   1. Run: cd ../supabase && supabase start"
    echo "   2. Copy JWT secret and anon key to .env"
    echo ""
    read -p "Press Enter when ready to continue..."
fi

# Check if Supabase is running
echo "🔍 Checking Supabase status..."
if ! nc -z localhost 54322 2>/dev/null; then
    echo "⚠️  Supabase doesn't appear to be running"
    echo "   Run: cd ../supabase && supabase start"
    read -p "Press Enter to continue anyway, or Ctrl+C to exit..."
else
    echo "✅ Supabase is running"
fi

echo ""
echo "🐳 Starting Docker services..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 5

echo ""
echo "✅ Services started!"
echo ""
echo "📊 Traefik Dashboard: http://localhost:8080"
echo "🌐 API Gateway: http://localhost"
echo ""
echo "Test endpoints:"
echo "  curl http://localhost/v1/cups"
echo "  curl http://localhost/v1/games"
echo ""
echo "View logs:"
echo "  docker-compose logs -f"
echo ""
echo "Stop services:"
echo "  docker-compose down"
echo ""


