# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DCV (Docker Compose Viewer) is a TUI tool for monitoring Docker Compose applications. It provides:
- List view of running Docker Compose applications with log viewing capability
- Special handling for dind (Docker-in-Docker) containers to view nested containers
- Vim-like navigation and commands throughout the interface

## Technical Architecture

- **Language**: Go (Golang)
- **TUI Framework**: tview
- **Core Functionality**: Wraps docker-compose commands to provide an interactive interface

## Key Views

1. **Process List View**: Shows `docker compose ps` results
   - Enter: View container logs
   - d: Treat as dind container and navigate to dind process list

2. **Dind Process List View**: Executes `docker ps` inside selected dind containers
   - Enter: View logs of containers running inside dind

3. **Log View**: Displays container logs with vim-like navigation
   - `/`: Search functionality
   - `G`: Jump to end
   - Standard vim navigation keys

## Development Guidelines

- Follow vim-style keybindings for all shortcuts
- The tool internally executes docker-compose commands
- Special handling required for dind (Docker-in-Docker) containers

## Build and Installation

The project is intended to be installed via `go install` (implementation pending).

## License

MIT License - Copyright © 2025 Tokuhiro Matsuno