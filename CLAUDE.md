# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DCV (Docker Container Viewer) is a TUI tool for monitoring Docker containers and Docker Compose applications. It provides:
- List view of all Docker containers (both plain Docker and Docker Compose managed)
- List and manage Docker images
- List and manage Docker networks
- List and manage Docker volumes
- Browse files inside containers
- Execute shell commands in containers
- Inspect container configuration
- Multiple Docker Compose project management and switching
- Log viewing capability for any container
- Special handling for dind (Docker-in-Docker) containers to view nested containers
- Vim-like navigation and command-line interface
- Built-in help system accessible with `?` key
- Quit confirmation dialog for safer exits
- Confirmation dialogs for aggressive operations (stop, start, kill, restart, delete, etc.)
- Search functionality in log views

## Technical Architecture

- **Language**: Go (Golang)
- **TUI Framework**: Bubble Tea (with Lipgloss for styling)
- **Architecture**: Model-View-Update (MVU) pattern
- **Core Functionality**: Wraps docker and docker-compose commands to provide an interactive interface
- **Configuration**: TOML-based configuration file support
- **Confirmation System**: Uses explicit `aggressive` boolean parameter in command execution to determine which operations require confirmation

## Key Views

1. **Docker Container List View**: Shows `docker ps` results for all containers
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: View container logs
   - `d`: Navigate to dind process list (for dind containers)
   - `f`: Browse container files
   - `!`: Execute /bin/sh in container
   - `I`: Inspect container (view full config)
   - `a`: Toggle show all containers (including stopped)
   - `s`: Show container stats
   - `K`: Kill container (with confirmation)
   - `S`: Stop container (with confirmation)
   - `U`: Start container (with confirmation)
   - `R`: Restart container (with confirmation)
   - `P`: Pause/Unpause container (with confirmation)
   - `D`: Delete stopped container (with confirmation)
   - `r`: Refresh list
   - `?`: Show help view
   - `:`: Enter command mode
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
   - `K`: Kill service (with confirmation)
   - `S`: Stop service (with confirmation)
   - `U`: Start service (with confirmation)
   - `R`: Restart service (with confirmation)
   - `P`: Pause/Unpause container (with confirmation)
   - `D`: Delete stopped container (with confirmation)
   - `u`: Deploy all services (docker compose up -d)
   - `x`: Stop and remove all services (docker compose down - with confirmation)
   - `r`: Refresh list
   - `?`: Show help view
   - `:`: Enter command mode
   - `q`: Quit (with confirmation)

3. **Image List View**: Shows Docker images with repository, tag, ID, creation time, and size
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `I`: Inspect image (view full config)
   - `a`: Toggle show all images (including intermediate layers)
   - `r`: Refresh list
   - `D`: Remove selected image (with confirmation)
   - `F`: Force remove selected image (even if used by containers - with confirmation)
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
   - `?`: Show help view
   - `Esc`/`q`: Back to process list

6. **Log View**: Displays container logs with vim-like navigation
   - `↑`/`k`: Scroll up
   - `↓`/`j`: Scroll down
   - `G`: Jump to end
   - `g`: Jump to start
   - `/`: Search logs (highlights matches in all logs)
   - `f`: Filter logs (shows only lines matching the filter)
   - `n`: Next search result
   - `N`: Previous search result
   - `?`: Show help view
   - `Esc`/`q`: Back to previous view

7. **Top View**: Shows process information (docker compose top)
   - `r`: Refresh
   - `?`: Show help view
   - `Esc`/`q`: Back to process list

8. **Stats View**: Shows container resource usage statistics
   - `r`: Refresh
   - `?`: Show help view
   - `Esc`/`q`: Back to process list

9. **Network List View**: Shows Docker networks with ID, name, driver, scope, and container count
   - `↑`/`k`: Move up
   - `↓`/`j`: Move down
   - `Enter`: Inspect network (view full config)
   - `r`: Refresh list
   - `D`: Remove selected network (except default networks - with confirmation)
   - `Esc`/`q`: Back to Docker Compose process list

10. **Volume List View**: Shows Docker volumes with name, driver, scope
    - `↑`/`k`: Move up
    - `↓`/`j`: Move down
    - `Enter`: Inspect volume (view full config)
    - `r`: Refresh list
    - `D`: Remove selected volume (with confirmation)
    - `F`: Force remove selected volume (with confirmation)
    - `Esc`/`q`: Back to Docker Compose process list

11. **File Browser View**: Browse filesystem inside containers
    - `↑`/`k`: Move up
    - `↓`/`j`: Move down
    - `Enter`: Open directory or view file
    - `u`: Go to parent directory
    - `r`: Refresh list
    - `?`: Show help view
    - `Esc`/`q`: Back to container list

12. **File Content View**: View file contents from containers
    - `↑`/`k`: Scroll up
    - `↓`/`j`: Scroll down
    - `G`: Jump to end
    - `g`: Jump to start
    - `?`: Show help view
    - `Esc`/`q`: Back to file browser

13. **Inspect View**: Shows full container configuration in JSON format
    - `↑`/`k`: Scroll up
    - `↓`/`j`: Scroll down
    - `G`: Jump to end
    - `g`: Jump to start
    - `?`: Show help view
    - `Esc`/`q`: Back to container list

14. **Help View**: Displays all available keyboard shortcuts for the current view
    - `↑`/`k`: Scroll up
    - `↓`/`j`: Scroll down
    - `Esc`/`q`: Back to previous view

15. **Command Mode**: Vim-style command line interface
    - `:q` or `:quit`: Quit with confirmation
    - `:q!` or `:quit!`: Force quit without confirmation
    - `:help commands`: List all available commands
    - `:set all`: Show all containers (including stopped)
    - `:set noall`: Hide stopped containers
    - **Key handler commands**: All key handlers can be called as commands (e.g., `:select-up-container`, `:show-compose-log`, `:kill-container`)
    - **Command aliases**: Common aliases like `:up`, `:down`, `:logs`, `:inspect`, `:exec`
    - `↑`/`↓`: Navigate command history
    - `←`/`→`: Move cursor in command line
    - `Enter`: Execute command
    - `Esc`: Cancel command mode

16. **Command Execution View**: Shows real-time output of Docker commands
    - Displays the exact Docker command being executed
    - Shows confirmation dialog for aggressive operations (stop, start, kill, restart, delete, etc.)
    - Confirmation dialog shows: "Are you sure you want to execute: docker [command]"
    - `y`: Confirm and execute the command
    - `n` or `Esc`: Cancel and return to previous view
    - `↑`/`k`: Scroll up through command output
    - `↓`/`j`: Scroll down through command output
    - `G`: Jump to end of output
    - `g`: Jump to start of output
    - `Ctrl+C`: Cancel running command
    - `Esc`: Return to previous view (after command completes)

## Development Guidelines

- Follow vim-style keybindings for all shortcuts
- The tool internally executes both docker and docker-compose commands
- Special handling required for dind (Docker-in-Docker) containers
- **Protected main branch**: The main branch is protected and cannot be committed to directly. Always create a feature branch and submit a pull request for code changes.
- **Confirmation System**: 
  - All command execution methods use `ExecuteCommand(model *Model, aggressive bool, args ...string)` signature
  - The `aggressive` boolean parameter explicitly determines if a confirmation dialog should be shown
  - Callers must explicitly specify whether operations are aggressive (stop, start, kill, restart, delete, etc.)
  - The confirmation dialog displays the exact Docker command that will be executed
  - No automatic detection or parsing - explicit parameter approach for maintainability

## Code Style and Quality

- **Code Formatting**: All code must be formatted with `goimports`
  - Run `make fmt` before committing
  - CI will fail if code is not properly formatted
- **Testing**: ALWAYS run tests before committing
  - Run `make fmt` first to ensure proper formatting
  - Run `make test` before every commit
  - All tests must pass before committing code
  - This is a mandatory step - no exceptions
- **Linting**: Code must pass all golangci-lint checks
  - Run `make lint` to check locally
  - Configuration is in `.golangci.yml`
- **Import Ordering**: 
  - Standard library imports first
  - Third-party imports second
  - Local imports last (with prefix `github.com/tokuhirom/dcv`)
- **Error Handling**: Always handle errors appropriately
- **Comments**: Add comments for exported functions and types

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
- **Image history** - Show layer history
- **Search Docker Hub** (`docker search`)
- **Build images** from Dockerfile

### Docker Compose Operations
- **Build services** (`docker compose build`)
- **Pull service images** (`docker compose pull`)
- **Scale services** (`docker compose scale`)
- **View compose file** - Display the active docker-compose.yml

### Network & Volume Management
- **Network connections** - Show container network relationships
- **Create networks/volumes** - Create new networks and volumes
- **Network/Volume usage visualization** - Show which containers use which networks/volumes

### UI/UX Enhancements
- **Search/filter** - Filter containers/images by name, status, etc.
- **Multiple selections** - Perform batch operations
- **Color themes** - Customizable color schemes
- **Export logs** - Save logs to file

### Monitoring Improvements
- **Real-time stats update** - Currently stats view doesn't auto-refresh
- **Resource usage graphs** - Visual representation of CPU/memory usage
- **Health check status** - Display container health status

### Configuration
- **Config file** - Save preferences (default view, filters, etc.)
- **Custom keybindings** - Allow users to customize shortcuts

## Configuration

DCV supports configuration through a TOML file located at:

- `~/.config/dcv/config.toml` - User config directory

### Configuration Options

```toml
[general]
# Initial view to show on startup
# Valid values: "docker", "compose", "projects"
# Default: "docker"
initial_view = "docker"
```

### Example Configuration

Copy `dcv.toml.example` to one of the locations above and modify as needed:

```bash
cp dcv.toml.example ~/.config/dcv/config.toml
```

## License

MIT License - Copyright © 2025 Tokuhiro Matsuno