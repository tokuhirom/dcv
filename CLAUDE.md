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

For a complete list of keyboard shortcuts and commands in each view, see [docs/keymap.md](docs/keymap.md).

1. **Docker Container List View**: Shows `docker ps` results for all containers

2. **Docker Compose Process List View**: Shows `docker compose ps` results

3. **Image List View**: Shows Docker images with repository, tag, ID, creation time, and size

4. **Project List View**: Shows all Docker Compose projects

5. **Dind Process List View**: Executes `docker ps` inside selected dind containers

6. **Log View**: Displays container logs with vim-like navigation
   - Supports search (`/`) and filter (`f`) functionality
   - Search highlights matches in all logs
   - Filter shows only lines matching the filter

7. **Top View**: Shows process information (docker compose top)

8. **Stats View**: Shows container resource usage statistics

9. **Network List View**: Shows Docker networks with ID, name, driver, scope, and container count

10. **Volume List View**: Shows Docker volumes with name, driver, scope

11. **File Browser View**: Browse filesystem inside containers

12. **File Content View**: View file contents from containers

13. **Inspect View**: Shows full container configuration in JSON format

14. **Help View**: Displays all available keyboard shortcuts for the current view

15. **Command Mode**: Vim-style command line interface
    - All key handlers can be called as commands
    - See [docs/keymap.md](docs/keymap.md#command-mode) for available commands

16. **Command Execution View**: Shows real-time output of Docker commands
    - Displays the exact Docker command being executed
    - Shows confirmation dialog for aggressive operations (stop, start, kill, restart, delete, etc.)
    - Confirmation dialog shows: "Are you sure you want to execute: docker [command]"

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
- **Screenshots**: When modifying UI code, screenshots must be updated
  - Run `go run -tags screenshots cmd/generate-screenshots/main.go` to generate screenshots
  - CI will fail if screenshots are not up-to-date
  - Lefthook will automatically generate screenshots on commit if UI files are changed

## Git Hooks with Lefthook

The project uses [lefthook](https://github.com/evilmartians/lefthook) for git hooks to ensure code quality:

- **Pre-commit hooks**:
  - Format code with `make fmt`
  - Run linter with `make lint`
  - Run tests with `make test`
  - Generate screenshots if UI files are modified
- **Pre-push hooks**:
  - Run full test suite

To set up lefthook:
```bash
# Install lefthook
go install github.com/evilmartians/lefthook@latest

# Install git hooks
lefthook install
```

The hooks configuration is in `lefthook.yml`

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

### Docker Compose Operations
- **Build services** (`docker compose build`)
- **Pull service images** (`docker compose pull`)
- **Scale services** (`docker compose scale`)
- **View compose file** - Display the active docker-compose.yml

### Network & Volume Management
- **Network connections** - Show container network relationships

### UI/UX Enhancements
- **Search/filter** - Filter containers/images by name, status, etc.
- **Multiple selections** - Perform batch operations
- **Color themes** - Customizable color schemes
- **Export logs** - Save logs to file

### Monitoring Improvements
- **Resource usage graphs** - Visual representation of CPU/memory usage

### Configuration
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

## Next Action Plan

Based on comprehensive codebase analysis, here are prioritized improvements:

### Immediate Actions (High Impact, Low Effort)
1. **Standardize Error Handling**
   - Create consistent error handling pattern across all views
   - Use message-based errors for UI updates
   - Return errors from Docker operations

2. **Extract Common View Model Base**
   - Create `BaseListViewModel` interface for all list-based views
   - Reduce code duplication across 16+ similar view models
   - Implement common navigation, search, and refresh logic

3. **Centralize Styling System**
   - Move all colors and styles to `internal/ui/styles.go`
   - Create theme support foundation
   - Eliminate hard-coded colors throughout codebase

4. **Add Command Timeouts**
   - Implement configurable timeouts for Docker commands
   - Prevent hanging on long-running operations
   - Add timeout configuration to TOML config

### Short Term Actions (1-2 weeks)
1. **Implement Copy Operations**
   - Add `docker cp` functionality for file transfers
   - Create UI for source/destination selection
   - Support both directions (to/from container)

2. **Add Bulk Operations**
   - Implement multi-select in container lists
   - Support batch operations (stop, start, remove)
   - Add select all/none shortcuts

3. **Improve Memory Management**
   - Add configurable limits for log storage
   - Implement ring buffer for large logs
   - Add pagination for large container/image lists

4. **Unified Confirmation Dialog**
   - Create single confirmation component
   - Standardize confirmation messages
   - Support custom confirmation prompts

### Medium Term Actions (2-4 weeks)
1. **Performance Optimization**
   - Implement lazy loading for large datasets
   - Add virtual scrolling for long lists
   - Cache frequently accessed data

2. **Advanced Search and Filtering**
   - Add persistent search/filter across views
   - Implement regex support in searches
   - Create filter presets

3. **Export Functionality**
   - Add ability to save logs to file
   - Export container lists as CSV/JSON
   - Support output redirection

4. **Resource Monitoring Enhancements**
   - Add real-time CPU/memory graphs
   - Implement historical data tracking
   - Create alert thresholds

### Long Term Actions (1-2 months)
1. **Plugin Architecture**
   - Design extensible command system
   - Support custom commands via plugins
   - Create plugin API documentation

2. **Remote Docker Support**
   - Add support for remote Docker hosts
   - Implement SSH tunnel support
   - Handle multiple Docker contexts

3. **Custom Key Bindings**
   - Create keybinding configuration system
   - Support user-defined shortcuts
   - Add keybinding conflict detection

4. **Advanced UI Features**
   - Implement split-pane views
   - Add tab support for multiple views
   - Create customizable layouts

## Code Quality Improvements

### Refactoring Priorities
1. **Interface Consolidation**
   - Merge single-method interfaces where appropriate
   - Create comprehensive view interfaces
   - Standardize interface naming conventions

2. **State Management**
   - Reduce Model struct complexity (currently 89 fields)
   - Implement view-specific state containers
   - Add state validation

3. **Test Coverage Enhancement**
   - Add integration tests for Docker operations
   - Improve error path testing
   - Create performance benchmarks
   - Add UI interaction tests

### Technical Debt Resolution
1. **View Model Duplication**
   - Extract common patterns to shared components
   - Use composition over inheritance
   - Implement view model factory pattern

2. **Naming Consistency**
   - Standardize view naming conventions
   - Unify function naming patterns
   - Consistent interface suffixes

3. **Documentation**
   - Add inline documentation for complex logic
   - Create architecture decision records (ADRs)
   - Document plugin development guide

## Implementation Strategy

1. **Phase 1 (Week 1-2)**: Foundation
   - Implement immediate actions
   - Set up base interfaces and styling system
   - Standardize error handling

2. **Phase 2 (Week 3-4)**: Core Features
   - Add copy operations and bulk selection
   - Implement memory management improvements
   - Create unified confirmation system

3. **Phase 3 (Week 5-8)**: Enhancement
   - Add export functionality
   - Implement advanced search/filtering
   - Optimize performance

4. **Phase 4 (Month 2+)**: Advanced Features
   - Design plugin architecture
   - Add remote Docker support
   - Implement custom keybindings

## Success Metrics
- Reduce code duplication by 40%
- Improve test coverage to 80%+
- Reduce memory usage for large datasets by 50%
- Decrease average response time by 30%
- Zero hanging operations with timeout implementation

## License

MIT License - Copyright © 2025 Tokuhiro Matsuno
