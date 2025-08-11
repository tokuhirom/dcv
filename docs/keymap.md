# DCV Keyboard Shortcuts and Commands

This document lists all keyboard shortcuts and commands available in DCV (Docker Container Viewer).

## Global Shortcuts

These shortcuts work across all views:

| Key | Description | Command |
|-----|-------------|----------|
| `q` | Quit | :quit |
| `:` | Command mode | :command-mode |
| `H` | Toggle navbar | :toggle-navbar |
| `1` | Docker container list | :ps |
| `2` | Project list | :compose-ls |
| `3` | Docker images | :images |
| `4` | Docker networks | :network-ls |
| `5` | Docker volumes | :volume-ls |
| `6` | Container stats | :stats |

## View-Specific Shortcuts

### Docker Compose Process List

View and manage Docker Compose containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `enter` | view logs | :view-logs |
| `esc` | back | :back |
| `?` | help | :help |
| `d` | entering DinD | :entering-dind |
| `x` | show actions | :show-actions |
| `f` | browse files | :browse-files |
| `!` | exec /bin/sh | :exec-/bin/sh |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `a` | toggle all | :toggle-all |
| `K` | kill | :kill |
| `S` | stop | :stop |
| `U` | start | :start |
| `R` | restart | :restart |
| `P` | pause/unpause | :pause/unpause |
| `D` | delete | :delete |
| `t` | top | :top |

### Docker Container List

View and manage all Docker containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `enter` | view logs | :view-logs |
| `esc` | back | :back |
| `?` | help | :help |
| `d` | entering DinD | :entering-dind |
| `x` | show actions | :show-actions |
| `f` | browse files | :browse-files |
| `!` | exec /bin/sh | :exec-/bin/sh |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `a` | toggle all | :toggle-all |
| `K` | kill | :kill |
| `S` | stop | :stop |
| `U` | start | :start |
| `R` | restart | :restart |
| `P` | pause/unpause | :pause/unpause |
| `D` | delete | :delete |
| `t` | top | :top |

### Log View

View container logs

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :scroll-up |
| `down, j` | scroll down | :scroll-down |
| `pgup` | page up | :page-up |
| `pgdown,  ` | page down | :page-down |
| `G` | go to end | :go-to-end |
| `g` | go to start | :go-to-start |
| `/` | search | :search |
| `n` | next match | :next-match |
| `N` | prev match | :prev-match |
| `f` | filter | :filter |
| `esc` | back | :back |
| `?` | help | :help |
| `ctrl+c` | cancel | :cancel |

### Top View

View container process information

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :scroll-up |
| `down, j` | scroll down | :scroll-down |
| `c` | sort by CPU | :sort-by-cpu |
| `m` | sort by memory | :sort-by-memory |
| `p` | sort by PID | :sort-by-pid |
| `t` | sort by time | :sort-by-time |
| `n` | sort by name | :sort-by-name |
| `R` | reverse sort | :reverse-sort |
| `a` | toggle auto-refresh | :toggle-auto-refresh |
| `r` | refresh | :refresh |
| `esc` | back | :back |
| `?` | help | :help |

### Stats View

View container resource statistics

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :scroll-up |
| `down, j` | scroll down | :scroll-down |
| `c` | sort by CPU | :sort-by-cpu |
| `m` | sort by memory | :sort-by-memory |
| `n` | sort by name | :sort-by-name |
| `R` | reverse sort | :reverse-sort |
| `a` | toggle auto-refresh | :toggle-auto-refresh |
| `r` | refresh | :refresh |
| `esc` | back | :back |
| `?` | help | :help |

### Project List

View and select Docker Compose projects

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `enter` | select project | :select-project |
| `r` | refresh | :refresh |
| `?` | help | :help |

### Image List

View and manage Docker images

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `a` | toggle all | :toggle-all |
| `D` | delete | :delete |
| `esc` | back | :back |
| `?` | help | :help |

### Network List

View and manage Docker networks

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `D` | delete | :delete |
| `esc` | back | :back |
| `?` | help | :help |

### Volume List

View and manage Docker volumes

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `D` | delete | :delete |
| `esc` | back | :back |
| `?` | help | :help |

### File Browser

Browse files inside containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `enter` | open | :open |
| `u` | parent directory | :parent-directory |
| `r` | refresh | :refresh |
| `esc` | back | :back |
| `?` | help | :help |

### File Content

View file contents from containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :scroll-up |
| `down, j` | scroll down | :scroll-down |
| `G` | go to end | :go-to-end |
| `g` | go to start | :go-to-start |
| `esc` | back | :back |
| `?` | help | :help |

### Inspect View

View detailed container/image/network/volume information

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :scroll-up |
| `down, j` | scroll down | :scroll-down |
| `pgup` | page up | :page-up |
| `pgdown,  ` | page down | :page-down |
| `G` | go to end | :go-to-end |
| `g` | go to start | :go-to-start |
| `/` | search | :search |
| `n` | next match | :next-match |
| `N` | prev match | :prev-match |
| `esc` | back | :back |
| `?` | help | :help |

### Help View

View help information

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :scroll-up |
| `down, j` | scroll down | :scroll-down |
| `esc` | back | :back |

### Docker in Docker

View containers inside dind containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :move-up |
| `down, j` | move down | :move-down |
| `enter` | view logs | :view-logs |
| `esc` | back | :back |
| `?` | help | :help |
| `x` | show actions | :show-actions |
| `f` | browse files | :browse-files |
| `!` | exec /bin/sh | :exec-/bin/sh |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `a` | toggle all | :toggle-all |
| `K` | kill | :kill |
| `S` | stop | :stop |
| `U` | start | :start |
| `R` | restart | :restart |
| `P` | pause/unpause | :pause/unpause |
| `D` | delete | :delete |
| `t` | top | :top |

### Command Execution

View command execution output

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :scroll-up |
| `down, j` | scroll down | :scroll-down |
| `G` | go to end | :go-to-end |
| `g` | go to start | :go-to-start |
| `ctrl+c` | cancel | :cancel |
| `esc` | back | :back |
| `?` | help | :help |

## Command Mode

Enter command mode by pressing `:`. Available commands:

| Command | Description |
|---------|-------------|
| `:q` or `:quit` | Quit DCV |
| `:q!` or `:quit!` | Force quit without confirmation |
| `:help commands` | List all available commands |
| `:set all` | Show all containers (including stopped) |
| `:set noall` | Hide stopped containers |

## Tips

- Most views support vim-style navigation (`j`/`k` for down/up)
- Press `?` in any view to see context-specific help
- Press `ESC` or `q` to go back to the previous view
- Press `H` to toggle the navigation bar visibility
- In process list views, press `x` to see available actions for a container

---
*This document is auto-generated. Do not edit manually.*
