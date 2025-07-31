# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DCV (Docker Compose Viewer) is a TUI tool for monitoring Docker Compose applications. It provides:
- List view of running Docker Compose applications with log viewing capability
- Special handling for dind (Docker-in-Docker) containers to view nested containers
- Vim-like navigation and commands throughout the interface

## Technical Architecture

- **Language**: Go (Golang)
- **TUI Framework**: Bubble Tea (with Lipgloss for styling)
- **Architecture**: Model-View-Update (MVU) pattern
- **Core Functionality**: Wraps docker-compose commands to provide an interactive interface

## Key Views

1. **Process List View**: Shows `docker compose ps` results
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View container logs
   - `d`: Navigate to dind process list (for dind containers)
   - `t`: Show process info (docker compose top)
   - `K`: Kill service (docker compose kill)
   - `S`: Stop service (docker compose stop)
   - `r`: Refresh list
   - `q`: Quit

2. **Dind Process List View**: Executes `docker ps` inside selected dind containers
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View logs of containers running inside dind
   - `r`: Refresh list
   - `Esc`/`q`: Back to process list

3. **Log View**: Displays container logs with vim-like navigation
   - `↑`/`k`: Scroll up
   - `↓`/`j`: Scroll down
   - `G`: Jump to end
   - `g`: Jump to start
   - `/`: Search functionality
   - `Esc`/`q`: Back to previous view

4. **Top View**: Shows process information (docker compose top)
   - `r`: Refresh
   - `Esc`/`q`: Back to process list

## Development Guidelines

- Follow vim-style keybindings for all shortcuts
- The tool internally executes docker-compose commands
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