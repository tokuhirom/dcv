#!/bin/bash
# Start script for DCV demo environment

set -e

echo "================================================"
echo "        DCV Demo Environment Startup"
echo "================================================"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null 2>&1; then
    echo "❌ docker-compose is not available. Please install it first."
    exit 1
fi

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Use docker compose if available, otherwise docker-compose
if docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

echo "🔧 Starting Docker Compose services..."
echo ""

# Start the services
$COMPOSE_CMD up -d

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 5

# Show status
echo ""
echo "📊 Service Status:"
echo "=================="
$COMPOSE_CMD ps

echo ""
echo "🐳 Docker in Docker containers will start shortly..."
echo "   Run 'docker exec -it demos-dind-1 docker ps' to see containers inside dind"
echo ""
echo "✅ Demo environment is ready!"
echo ""
echo "🎯 Quick Commands:"
echo "  • View all containers:     dcv"
echo "  • View dind containers:    docker exec -it demos-dind-1 docker ps"
echo "  • View logs:              $COMPOSE_CMD logs -f [service]"
echo "  • Stop all:               $COMPOSE_CMD down"
echo "  • Stop and remove data:   $COMPOSE_CMD down -v"
echo ""
echo "📝 Services running:"
echo "  • redis        - Redis cache server (Alpine)"
echo "  • dind         - Docker in Docker with lightweight containers"
echo "  • logger       - Verbose logging service"
echo "  • cache-warmer - Cache warming service"
echo ""
echo "🔍 Inside DinD container:"
echo "  • echo-server      - HTTP echo server"
echo "  • web-server       - Nginx web server (Alpine)"
echo "  • cache           - Redis cache (Alpine)"
echo "  • worker-1/2      - Background workers"
echo "  • healthcheck-demo - Container with health checks"
echo ""