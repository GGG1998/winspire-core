#!/bin/bash
# Stop script for local development

set -e

echo "Stopping Winspire Local Development Environment"
echo ""

if [ ! -f "docker-compose.yaml" ]; then
    echo "Error: Must be run from platform/local/ directory"
    exit 1
fi

# Stop all services including scaled ones
docker-compose down
docker-compose --profile scale down 2>/dev/null || true

echo ""
echo "All services stopped"
echo ""
echo "To start again: ./start.sh"
echo "To clean volumes: docker-compose down -v"
echo "To clean everything: make clean"
echo ""
