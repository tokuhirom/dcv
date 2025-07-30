# DCV Test Environment

This directory contains a test Docker Compose setup for testing DCV (Docker Compose Viewer).

## Quick Start

1. Start the test environment:
```bash
make up
```

2. Run DCV:
```bash
./dcv
```

Or build and run in one command:
```bash
make run
```

## Test Environment

The docker-compose.yml sets up:
- **web**: Nginx serving a test page on port 8090
- **db**: PostgreSQL database
- **redis**: Redis cache server
- **dind**: Docker-in-Docker container running:
  - echo-server: Simple HTTP echo server
  - test-db: MariaDB database
  - worker-1: Background worker with job logs
  - worker-2: Another worker with different logs

## Commands

- `make up` - Start all services and setup dind containers
- `make down` - Stop and remove all services
- `make logs` - Show logs from all services
- `make test-dind` - Manually setup containers in dind
- `make clean` - Remove everything including volumes
- `make run` - Build dcv and run with test environment

## Testing DCV Features

1. **View all containers**: The main screen shows all docker-compose services
2. **View logs**: Press Enter on any container to see its logs
3. **Dind support**: Press 'd' on the dind container to see containers running inside it
4. **Vim navigation**: Use j/k to navigate, G to go to end, / to search in logs
5. **Exit**: Press 'q' or Esc to go back/quit