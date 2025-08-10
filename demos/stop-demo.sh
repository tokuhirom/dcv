#!/bin/bash
# Stop script for DCV demo environment

set -e

echo "================================================"
echo "        DCV Demo Environment Shutdown"
echo "================================================"
echo ""

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Use docker compose if available, otherwise docker-compose
if docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

echo "🛑 Stopping Docker Compose services..."
$COMPOSE_CMD down

echo ""
echo "✅ Demo environment stopped!"
echo ""
echo "💡 Tips:"
echo "  • To remove volumes too: $COMPOSE_CMD down -v"
echo "  • To restart: ./start-demo.sh"
echo ""