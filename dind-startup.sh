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
echo "Starting test containers inside dind..."

# 1. Simple echo server
docker run -d \
    --name echo-server \
    --restart unless-stopped \
    -p 8000:8000 \
    hashicorp/http-echo:latest \
    -text="Hello from DCV test environment!"

# 2. Test database
docker run -d \
    --name test-mysql \
    --restart unless-stopped \
    -e MYSQL_ROOT_PASSWORD=rootpass \
    -e MYSQL_DATABASE=testdb \
    -e MYSQL_USER=testuser \
    -e MYSQL_PASSWORD=testpass \
    mysql:5.7

# 3. Background worker
docker run -d \
    --name worker-1 \
    --restart unless-stopped \
    alpine:latest \
    sh -c 'while true; do echo "[$(date)] Processing job #$RANDOM"; sleep 5; done'

# 4. Another worker with different logs
docker run -d \
    --name worker-2 \
    --restart unless-stopped \
    alpine:latest \
    sh -c 'while true; do echo "[$(date)] Worker 2: Task completed successfully"; sleep 8; done'

echo "All test containers started successfully!"
docker ps