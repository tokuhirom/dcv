# dcv - Docker Container Viewer

DCV is a TUI (Terminal User Interface) tool for monitoring Docker containers and Docker Compose applications.

## Features

- View all Docker containers (both standalone and Docker Compose managed)
- List and manage Docker images
- List and manage Docker networks
- List and manage Docker volumes with size information
- Browse files inside containers
- Execute shell commands in containers
- Inspect container configuration
- List and manage Docker Compose containers
- Switch between multiple Docker Compose projects
- Real-time container log streaming (shows last 1000 lines, then follows new logs)
- Manage containers inside Docker-in-Docker (dind) containers
- Vim-style key bindings and command-line interface
- Help view accessible with `?` key
- Quit confirmation dialog for safer exits
- Display executed commands for debugging

## Views

### Docker Container List View

Displays `docker ps` results in a table format. Shows all Docker containers, not limited to Docker Compose.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down
- `Enter`: View container logs
- `f`: Browse container files
- `!`: Execute /bin/sh in container
- `I`: Inspect container (view full config)
- `a`: Toggle show all containers (including stopped)
- `s`: Show container stats
- `r`: Refresh list
- `K`: Kill container
- `S`: Stop container
- `U`: Start container
- `R`: Restart container
- `P`: Pause/Unpause container
- `D`: Delete stopped container
- `?`: Show help view with all keybindings
- `:`: Enter command mode (vim-style commands)
- `q`/`Esc`: Back to Docker Compose process list

### Docker Compose Process List View

Displays `docker compose ps` results in a table format.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down  
- `Enter`: View container logs
- `d`: View dind container contents (only for dind containers)
- `f`: Browse container files
- `!`: Execute /bin/sh in container
- `I`: Inspect container (view full config)
- `p`: Switch to Docker container list view
- `i`: Switch to Docker image list view
- `n`: Switch to Docker network list view
- `P`: Switch to project list view
- `V`: Switch to volume list view
- `1`: Quick switch to Docker container list view
- `2`: Quick switch to project list view
- `3`: Quick switch to Docker image list view
- `4`: Quick switch to Docker network list view
- `5`: Quick switch to Docker volume list view
- `a`: Toggle show all containers (including stopped)
- `r`: Refresh list
- `t`: Show process info (docker compose top)
- `s`: Show container stats
- `K`: Kill service
- `S`: Stop service
- `U`: Start service
- `R`: Restart service
- `P`: Pause/Unpause container
- `D`: Delete stopped container
- `u`: Deploy all services (docker compose up -d)
- `x`: Stop and remove all services (docker compose down)
- `?`: Show help view with all keybindings
- `:`: Enter command mode (vim-style commands)
- `q`: Quit (with confirmation)

### Log View

Displays container logs. Initially shows the last 1000 lines, then streams new logs in real-time.

**Key bindings:**
- `↑`/`k`: Scroll up
- `↓`/`j`: Scroll down
- `G`: Jump to end
- `g`: Jump to start
- `/`: Search logs (supports regex with Ctrl+R, case insensitive with Ctrl+I)
- `f`: Filter logs (shows only lines matching the filter)
- `n`: Next search result
- `N`: Previous search result
- `?`: Show help view
- `Esc`/`q`: Back to previous view

### Docker-in-Docker Process List View

Shows containers running inside a dind container.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down
- `Enter`: View container logs
- `r`: Refresh list
- `?`: Show help view
- `Esc`/`q`: Back to process list

### Image List View

Displays Docker images with repository, tag, ID, creation time, and size information.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down
- `I`: Inspect image (view full config)
- `a`: Toggle show all images (including intermediate layers)
- `r`: Refresh list
- `D`: Remove selected image
- `F`: Force remove selected image (even if used by containers)
- `?`: Show help view
- `Esc`/`q`: Back to Docker Compose process list

### Network List View

Displays Docker networks with ID, name, driver, scope, and container count.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down
- `Enter`: Inspect network (view full config)
- `r`: Refresh list
- `D`: Remove selected network (except default networks)
- `?`: Show help view
- `Esc`/`q`: Back to Docker Compose process list

### Volume List View

Displays Docker volumes with name, driver, scope, size, creation time, and reference count.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down
- `Enter`: Inspect volume (view full config)
- `r`: Refresh list
- `D`: Remove selected volume
- `F`: Force remove selected volume (even if used by containers)
- `?`: Show help view
- `Esc`/`q`: Back to Docker Compose process list

### File Browser View

Browse the filesystem inside a container. Navigate directories and view file contents.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down
- `Enter`: Open directory or view file
- `u`: Go to parent directory
- `r`: Refresh list
- `?`: Show help view
- `Esc`/`q`: Back to container list

### File Content View

View the contents of a file from within a container.

**Key bindings:**
- `↑`/`k`: Scroll up
- `↓`/`j`: Scroll down
- `G`: Jump to end
- `g`: Jump to start
- `?`: Show help view
- `Esc`/`q`: Back to file browser

### Inspect View

Displays the full Docker inspect output for containers, images, or networks in JSON format with syntax highlighting.

**Key bindings:**
- `↑`/`k`: Scroll up
- `↓`/`j`: Scroll down
- `G`: Jump to end
- `g`: Jump to start
- `/`: Search
- `n`: Next search result
- `N`: Previous search result
- `?`: Show help view
- `Esc`/`q`: Back to previous view

### Top View

Shows process information (docker compose top) for the selected container.

**Key bindings:**
- `r`: Refresh
- `?`: Show help view
- `Esc`/`q`: Back to process list

### Stats View

Shows container resource usage statistics including CPU, memory, network I/O, and block I/O.

**Key bindings:**
- `r`: Refresh
- `?`: Show help view
- `Esc`/`q`: Back to process list

### Compose Project List View

Shows all Docker Compose projects on the system.

**Key bindings:**
- `↑`/`k`: Move up
- `↓`/`j`: Move down
- `Enter`: Select project and view its containers
- `r`: Refresh list
- `?`: Show help view
- `q`: Quit

### Help View

Shows all available keyboard shortcuts and their corresponding commands for the current view.

The help view displays three columns:
- **Key**: The keyboard shortcut (e.g., 'j' or 'down')
- **Command**: The vim-style command that can be used in command mode (e.g., ':cmd-down')
- **Description**: What the action does

**Key bindings:**
- `↑`/`k`: Scroll up
- `↓`/`j`: Scroll down
- `Esc`/`q`: Back to previous view

### Command Mode

Vim-style command line interface for executing commands.

**Built-in commands:**
- `:q` or `:quit`: Quit the application (with confirmation)
- `:q!` or `:quit!`: Force quit without confirmation
- `:h` or `:help`: Show help view
- `:help commands`: List all available commands in current view

**Key handler commands:**
All key handler functions can be called as commands. Commands are automatically derived from the handler method names and use either short forms or kebab-case naming:
- `:up` or `:cmd-up`: Move selection up
- `:down` or `:cmd-down`: Move selection down
- `:log` or `:cmd-log`: View container logs
- `:kill` or `:cmd-kill`: Kill selected container
- `:stop` or `:cmd-stop`: Stop selected container
- `:start` or `:cmd-start`: Start selected container
- `:restart` or `:cmd-restart`: Restart selected container
- `:filebrowse` or `:cmd-file-browse`: Browse container files
- `:shell` or `:cmd-shell`: Execute /bin/sh in container
- `:inspect` or `:cmd-inspect`: Show inspect view
- `:stats` or `:cmd-stats`: Show container stats
- `:top` or `:cmd-top`: Show process info
- And many more...

**Command aliases:**
- `:up` → Move selection up
- `:down` → Move selection down
- `:log` or `:logs` → View container logs
- `:inspect` → Show inspect view
- `:shell` or `:exec` → Execute shell in container
- `:ps` → Show Docker container list
- `:images` → Show image list
- `:networks` → Show network list
- `:volumes` → Show volume list
- `:remove` or `:rm` → Remove/delete container
- `:all` or `:a` → Toggle show all containers
- `:refresh` or `:r` → Refresh current view
- `:quit` or `:q` → Quit application
- `:help` or `:h` → Show help

**Key bindings:**
- `:`: Enter command mode from any view
- `↑`/`↓`: Navigate command history
- `←`/`→`: Move cursor in command line
- `Enter`: Execute command
- `Esc`: Cancel command mode

## Usage

### Options

```bash
dcv [-p <project>] [-f <compose-file>] [--projects]
```

- `-p <project>`: Display the specified Docker Compose project
- `-f <compose-file>`: Display the project with the specified compose file
- `--projects`: Show project list on startup

### Examples

```bash
# Display Docker Compose project in current directory
dcv

# Display specific project
dcv -p myproject

# Start with project list
dcv --projects

# To view all Docker containers, press 'p' after starting
```

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

```bash
# Create config directory
mkdir -p ~/.config/dcv

# Create config file
cat > ~/.config/dcv/config.toml << EOF
[general]
initial_view = "compose"
EOF
```

## Installation

### Using go install

```bash
go install github.com/tokuhirom/dcv@latest
```

### Building from source

```bash
git clone https://github.com/tokuhirom/dcv.git
cd dcv
go build -o dcv
./dcv
```

## Requirements

- Go 1.23 or later (for building from source)
- Docker and Docker Compose installed
- Terminal with TUI support

## Implementation Details

- Language: Go
- TUI Framework: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styling: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Testing: testify

### Architecture

- Uses Model-View-Update (MVU) pattern
- Async log streaming
- Shows executed commands on error for easier debugging
- Comprehensive unit tests

## Command Execution View

Shows real-time output when executing Docker commands (start, stop, restart, kill, rm, docker compose up/down).

**Key bindings:**
- `↑`/`k`: Scroll up
- `↓`/`j`: Scroll down
- `G`: Jump to end
- `g`: Jump to start
- `Ctrl+C`: Cancel running command
- `Esc`: Back to previous view

## Debugging Features

- Commands are displayed when executing operations
- Real-time streaming output for all Docker operations
- Error messages with exit codes
- Color-coded status indicators (success/failure)

## Development

### Setting up development environment

```bash
# Install all development dependencies including lefthook
make dev-deps
```

This will install:
- golangci-lint for linting
- goimports for code formatting
- lefthook for git hooks

### Git Hooks (Lefthook)

This project uses [lefthook](https://github.com/evilmartians/lefthook) to ensure code quality:

- **Pre-commit**: Automatically formats Go code and runs quick tests
- **Pre-push**: Runs full linting and test suite

To manually run hooks:
```bash
lefthook run pre-commit
lefthook run pre-push
```

### Running tests

```bash
make test
```

### Building

```bash
make all
```

### Formatting code

```bash
make fmt
```


## License

```
The MIT License (MIT)

Copyright © 2025 Tokuhiro Matsuno, https://64p.org/ <tokuhirom@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```