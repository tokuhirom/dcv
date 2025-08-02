# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DCV (Docker Container Viewer) is a TUI tool for monitoring Docker containers and Docker Compose applications. It provides:
- List view of all Docker containers (both plain Docker and Docker Compose managed)
- List and manage Docker images
- List and manage Docker networks
- Browse files inside containers
- Execute shell commands in containers
- Inspect container configuration
- Multiple Docker Compose project management and switching
- Log viewing capability for any container
- Special handling for dind (Docker-in-Docker) containers to view nested containers
- Vim-like navigation and commands throughout the interface

## Technical Architecture

- **Language**: Go (Golang)
- **TUI Framework**: Bubble Tea (with Lipgloss for styling)
- **Architecture**: Model-View-Update (MVU) pattern
- **Core Functionality**: Wraps docker and docker-compose commands to provide an interactive interface

## Key Views

1. **Docker Container List View**: Shows `docker ps` results for all containers
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View container logs
   - `f`: Browse container files
   - `!`: Execute /bin/sh in container
   - `I`: Inspect container (view full config)
   - `a`: Toggle show all containers (including stopped)
   - `K`: Kill container
   - `S`: Stop container
   - `U`: Start container
   - `R`: Restart container
   - `P`: Pause/Unpause container
   - `D`: Delete stopped container
   - `r`: Refresh list
   - `q`/`Esc`: Back to Docker Compose view

2. **Docker Compose Process List View**: Shows `docker compose ps` results
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View container logs
   - `d`: Navigate to dind process list (for dind containers)
   - `f`: Browse container files
   - `!`: Execute /bin/sh in container
   - `I`: Inspect container (view full config)
   - `p`: Switch to Docker container list view
   - `i`: Switch to Docker image list view
   - `n`: Switch to Docker network list view
   - `P`: Switch to project list view
   - `a`: Toggle show all containers (including stopped)
   - `t`: Show process info (docker compose top)
   - `s`: Show container stats
   - `K`: Kill service
   - `S`: Stop service
   - `U`: Start service
   - `R`: Restart service
   - `P`: Pause/Unpause container
   - `D`: Delete stopped container
   - `u`: Deploy all services (docker compose up -d)
   - `r`: Refresh list
   - `q`: Quit

3. **Image List View**: Shows Docker images with repository, tag, ID, creation time, and size
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `a`: Toggle show all images (including intermediate layers)
   - `r`: Refresh list
   - `D`: Remove selected image
   - `F`: Force remove selected image (even if used by containers)
   - `Esc`/`q`: Back to Docker Compose process list

4. **Project List View**: Shows all Docker Compose projects
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: Select project and view its containers
   - `r`: Refresh list
   - `q`: Quit

5. **Dind Process List View**: Executes `docker ps` inside selected dind containers
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View logs of containers running inside dind
   - `r`: Refresh list
   - `Esc`/`q`: Back to process list

6. **Log View**: Displays container logs with vim-like navigation
   - `↑`/`k`: Scroll up
   - `↓`/`j`: Scroll down
   - `G`: Jump to end
   - `g`: Jump to start
   - `/`: Search functionality
   - `Esc`/`q`: Back to previous view

7. **Top View**: Shows process information (docker compose top)
   - `r`: Refresh
   - `Esc`/`q`: Back to process list

8. **Stats View**: Shows container resource usage statistics
   - `r`: Refresh
   - `Esc`/`q`: Back to process list

9. **Network List View**: Shows Docker networks with ID, name, driver, scope, and container count
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `r`: Refresh list
   - `D`: Remove selected network (except default networks)
   - `Esc`/`q`: Back to Docker Compose process list

10. **File Browser View**: Browse filesystem inside containers
    - `↑`/`k`: Move up
    - `↓`/`j`: Move down
    - `Enter`: Open directory or view file
    - `r`: Refresh list
    - `Esc`/`q`: Back to container list

11. **File Content View**: View file contents from containers
    - `↑`/`k`: Scroll up
    - `↓`/`j`: Scroll down
    - `G`: Jump to end
    - `g`: Jump to start
    - `Esc`/`q`: Back to file browser

12. **Inspect View**: Shows full container configuration in JSON format
    - `↑`/`k`: Scroll up
    - `↓`/`j`: Scroll down
    - `G`: Jump to end
    - `g`: Jump to start
    - `Esc`/`q`: Back to container list

## Development Guidelines

- Follow vim-style keybindings for all shortcuts
- The tool internally executes both docker and docker-compose commands
- Special handling required for dind (Docker-in-Docker) containers

## Build and Installation

```bash
# Install via go install
go install github.com/tokuhirom/dcv@latest

# Or build from source
git clone https://github.com/tokuhirom/dcv.git
cd dcv
go build -o dcv
```

## Potential Missing Features

### Container Management
- **Copy files to/from containers** (`docker cp`)
- **Container rename** - Change container names
- **Download files from container** - Save container files locally

### Image Management
- **Pull images** (`docker pull`) - Download new images
- **Image inspect** - View detailed image information
- **Image history** - Show layer history
- **Search Docker Hub** (`docker search`)
- **Build images** from Dockerfile

### Docker Compose Operations
- **Down command** (`docker compose down`) - Stop and remove containers
- **Build services** (`docker compose build`)
- **Pull service images** (`docker compose pull`)
- **Scale services** (`docker compose scale`)
- **View compose file** - Display the active docker-compose.yml

### Network & Volume Management
- **Network inspect** - View detailed network information
- **Volume list/inspect** - Manage Docker volumes
- **Network connections** - Show container network relationships

### UI/UX Enhancements
- **Search/filter** - Filter containers/images by name, status, etc.
- **Multiple selections** - Perform batch operations
- **Color themes** - Customizable color schemes
- **Export logs** - Save logs to file
- **Log search** - The `/` search in log view is marked as "not implemented yet"

### Monitoring Improvements
- **Real-time stats update** - Currently stats view doesn't auto-refresh
- **Resource usage graphs** - Visual representation of CPU/memory usage
- **Health check status** - Display container health status

### Configuration
- **Config file** - Save preferences (default view, filters, etc.)
- **Custom keybindings** - Allow users to customize shortcuts

## License

MIT License - Copyright © 2025 Tokuhiro Matsuno