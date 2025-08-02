# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DCV (Docker Container Viewer) is a TUI tool for monitoring Docker containers and Docker Compose applications. It provides:
- List view of all Docker containers (both plain Docker and Docker Compose managed)
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
   - `a`: Toggle show all containers (including stopped)
   - `K`: Kill container
   - `S`: Stop container
   - `U`: Start container
   - `R`: Restart container
   - `D`: Delete stopped container
   - `r`: Refresh list
   - `q`/`Esc`: Back to Docker Compose view

2. **Docker Compose Process List View**: Shows `docker compose ps` results
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View container logs
   - `d`: Navigate to dind process list (for dind containers)
   - `p`: Switch to Docker container list view
   - `P`: Switch to project list view
   - `t`: Show process info (docker compose top)
   - `K`: Kill service
   - `S`: Stop service
   - `r`: Refresh list
   - `q`: Quit

3. **Project List View**: Shows all Docker Compose projects
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: Select project and view its containers
   - `r`: Refresh list
   - `q`: Quit

4. **Dind Process List View**: Executes `docker ps` inside selected dind containers
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View logs of containers running inside dind
   - `r`: Refresh list
   - `Esc`/`q`: Back to process list

5. **Log View**: Displays container logs with vim-like navigation
   - `↑`/`k`: Scroll up
   - `↓`/`j`: Scroll down
   - `G`: Jump to end
   - `g`: Jump to start
   - `/`: Search functionality
   - `Esc`/`q`: Back to previous view

6. **Top View**: Shows process information (docker compose top)
   - `r`: Refresh
   - `Esc`/`q`: Back to process list

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

## License

MIT License - Copyright © 2025 Tokuhiro Matsuno