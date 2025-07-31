#!/bin/sh
# Custom entrypoint for dind container

# Start dockerd in the background
dockerd-entrypoint.sh "$@" &
DOCKERD_PID=$!

# Wait for Docker daemon to be ready and run startup script
(
    max_attempts=30
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if docker info >/dev/null 2>&1; then
            echo "Docker daemon is ready, running startup script..."
            sh /startup.sh
            break
        fi
        echo "Waiting for Docker daemon... (attempt $((attempt+1))/$max_attempts)"
        sleep 1
        attempt=$((attempt+1))
    done
) &

# Wait for dockerd
wait $DOCKERD_PID