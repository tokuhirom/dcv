# DCV Demo Environment

This directory contains a Docker Compose demo environment for testing and demonstrating DCV (Docker Container Viewer) features.

## Quick Start

```bash
# Start the demo environment
docker compose up -d

# View with DCV from the parent directory
cd ..
./dcv

# Stop the demo environment
docker compose down
```

## Services

The demo environment includes several services to showcase DCV's capabilities:

### Redis Cache
- **Service**: `redis`
- **Image**: `redis:8-alpine`
- **Purpose**: Demonstrates container management and volume persistence
- **Network**: backend

### Docker-in-Docker (dind)
- **Service**: `dind`
- **Image**: `docker:28-dind`
- **Purpose**: Demonstrates DCV's ability to view containers inside dind containers
- **Features**: Automatically starts test containers on startup
- **Network**: development (isolated)

### Logger Service
- **Service**: `logger`
- **Image**: `alpine:latest`
- **Purpose**: Generates continuous log output for testing log viewing features
- **Output**: Timestamps, system info, memory usage, and random data every 2 seconds
- **Network**: monitoring

### Cache Warmer
- **Service**: `cache-warmer`
- **Image**: `alpine:latest`
- **Purpose**: Simple service that runs periodically
- **Network**: backend

## Networks

The demo uses multiple networks to demonstrate DCV's network viewing capabilities:

- **dcv-main-network** (172.24.0.0/16): Default bridge network
- **dcv-frontend** (172.25.0.0/16): For web-facing services
- **dcv-backend** (172.26.0.0/16): For database and cache services
- **dcv-monitoring** (172.27.0.0/16): For monitoring services
- **dcv-development** (172.28.0.0/16): Internal network for development tools

## Volumes

Persistent volumes used in the demo:

- **postgres-data**: Database storage (if postgres is added)
- **redis-data**: Redis append-only file storage
- **dind-storage**: Docker-in-Docker storage

## Testing DCV Features

### Container Management
1. Start the demo environment
2. Launch DCV to see all running containers
3. Try stopping, starting, and restarting services
4. Use the kill command to force-stop containers
5. Delete stopped containers

### Log Viewing
1. Select the `logger` service and press Enter to view logs
2. Use `/` to search through logs
3. Use `f` to filter logs
4. Press `G` to jump to the end of logs

### Docker-in-Docker
1. Select the `dind` container
2. Press `d` to view containers running inside dind
3. The dind container automatically starts test containers on startup

### Network Management
1. Press `n` to view the network list
2. See the multiple networks created by the demo
3. Inspect network details with Enter

### Volume Management
1. Press `V` or `5` to view the volume list
2. See the persistent volumes created by the demo
3. Inspect volume details with Enter

### File Browsing
1. Select any container and press `f` to browse its filesystem
2. Navigate directories and view file contents
3. Use `u` to go up to parent directory

## Cleanup

To completely remove the demo environment including volumes:

```bash
docker compose down -v
```

This will remove:
- All containers
- All networks created by the demo
- All volumes (data will be lost)

## Customization

Feel free to modify `docker-compose.yml` to add more services or change configurations to test different DCV features.