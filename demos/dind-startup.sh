#!/bin/sh
# Startup script for dind container

# Wait for Docker daemon
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if docker info >/dev/null 2>&1; then
        echo "Docker daemon is ready!"
        break
    fi
    echo "Waiting for Docker daemon... (attempt $((attempt+1))/$max_attempts)"
    sleep 1
    attempt=$((attempt+1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "Docker daemon failed to start!"
    exit 1
fi

# Start test containers
echo "Starting lightweight test containers inside dind..."

# Check if containers already exist and remove them if they do
for container in echo-server web-server cache worker-1 worker-2 healthcheck-demo; do
    if docker ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
        echo "Removing existing container: $container"
        docker rm -f "$container" 2>/dev/null || true
    fi
done

# 1. Simple HTTP echo server (very lightweight)
docker run -d \
    --name echo-server \
    --restart unless-stopped \
    -p 8000:8000 \
    hashicorp/http-echo:latest \
    -text="Hello from DCV test environment!"

# 2. Nginx web server (lightweight alternative to database)
docker run -d \
    --name web-server \
    --restart unless-stopped \
    -p 8080:80 \
    nginx:alpine

# 3. Redis cache (much lighter than MariaDB)
docker run -d \
    --name cache \
    --restart unless-stopped \
    -p 6379:6379 \
    redis:alpine \
    redis-server --appendonly yes

# 4. Background worker
docker run -d \
    --name worker-1 \
    --restart unless-stopped \
    alpine:latest \
    sh -c 'while true; do echo "[$(date)] Processing job #$RANDOM"; sleep 5; done'

# 5. Another worker with different logs
docker run -d \
    --name worker-2 \
    --restart unless-stopped \
    alpine:latest \
    sh -c 'while true; do echo "[$(date)] Worker 2: Task completed successfully"; sleep 8; done'

# 6. BusyBox health check simulator
docker run -d \
    --name healthcheck-demo \
    --restart unless-stopped \
    --health-cmd="echo 'Health check OK'" \
    --health-interval=30s \
    --health-timeout=3s \
    --health-retries=3 \
    busybox:latest \
    sh -c 'while true; do echo "[$(date)] Health check demo running..."; sleep 10; done'

echo "All test containers started successfully!"
docker ps