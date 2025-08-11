# DCV Keyboard Shortcuts and Commands

This document lists all keyboard shortcuts and commands available in DCV (Docker Container Viewer).

## Global Shortcuts

These shortcuts work across all views:

| Key | Description | Command |
|-----|-------------|----------|
| `q` | quit | :quit |
| `:` | command mode | :command-mode |
| `H` | toggle navbar | :toggle-navbar |
| `1` | docker ps | :ps |
| `2` | project list | :compose-ls |
| `3` | docker images | :images |
| `4` | docker networks | :network-ls |
| `5` | docker volumes | :volume-ls |
| `6` | stats | :stats |

## View-Specific Shortcuts

### Docker Compose Process List

View and manage Docker Compose containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :up |
| `down, j` | move down | :down |
| `enter` | view logs | :log |
| `esc` | back | :back |
| `?` | help | :help |
| `d` | entering DinD | :dind |
| `x` | show actions | :show-actions |
| `f` | browse files | :file-browse |
| `!` | exec /bin/sh | :shell |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `a` | toggle all | :toggle-all |
| `K` | kill | :kill |
| `S` | stop | :stop |
| `U` | start | :start |
| `R` | restart | :restart |
| `P` | pause/unpause | :pause |
| `D` | delete | :delete |
| `t` | top | :top |

### Docker Container List

View and manage all Docker containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :up |
| `down, j` | move down | :down |
| `enter` | view logs | :log |
| `esc` | back | :back |
| `?` | help | :help |
| `d` | entering DinD | :dind |
| `x` | show actions | :show-actions |
| `f` | browse files | :file-browse |
| `!` | exec /bin/sh | :shell |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `a` | toggle all | :toggle-all |
| `K` | kill | :kill |
| `S` | stop | :stop |
| `U` | start | :start |
| `R` | restart | :restart |
| `P` | pause/unpause | :pause |
| `D` | delete | :delete |
| `t` | top | :top |

### Log View

View container logs

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :up |
| `down, j` | scroll down | :down |
| `pgup` | page up | :page-up |
| `pgdown,  ` | page down | :page-down |
| `G` | go to end | :go-to-end |
| `g` | go to start | :go-to-start |
| `/` | search | :search |
| `n` | next match | :next-search-result |
| `N` | prev match | :prev-search-result |
| `f` | filter | :filter |
| `esc` | back | :back |
| `?` | help | :help |
| `ctrl+c` | cancel | :cancel |

### Top View

View container process information

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :up |
| `down, j` | scroll down | :down |
| `c` | sort by CPU | :sort-by-cpu |
| `m` | sort by memory | :sort-by-mem |
| `p` | sort by PID | :sort-by-pid |
| `t` | sort by time | :sort-by-time |
| `n` | sort by name | :sort-by-command |
| `R` | reverse sort | :reverse-sort |
| `a` | toggle auto-refresh | :toggle-auto-refresh |
| `r` | refresh | :refresh |
| `esc` | back | :back |
| `?` | help | :help |

### Stats View

View container resource statistics

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :up |
| `down, j` | scroll down | :down |
| `c` | sort by CPU | :sort-by-cpu |
| `m` | sort by memory | :sort-by-mem |
| `n` | sort by name | :sort-by-command |
| `R` | reverse sort | :reverse-sort |
| `a` | toggle auto-refresh | :toggle-auto-refresh |
| `r` | refresh | :refresh |
| `esc` | back | :back |
| `?` | help | :help |

### Project List

View and select Docker Compose projects

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :up |
| `down, j` | move down | :down |
| `enter` | select project | :select-project |
| `r` | refresh | :refresh |
| `?` | help | :help |

### Image List

View and manage Docker images

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :up |
| `down, j` | move down | :down |
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
| `up, k` | move up | :up |
| `down, j` | move down | :down |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `D` | delete | :delete |
| `esc` | back | :back |
| `?` | help | :help |

### Volume List

View and manage Docker volumes

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :up |
| `down, j` | move down | :down |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `D` | delete | :delete |
| `esc` | back | :back |
| `?` | help | :help |

### File Browser

Browse files inside containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :up |
| `down, j` | move down | :down |
| `enter` | open | :open-file-or-directory |
| `u` | parent directory | :go-to-parent-directory |
| `r` | refresh | :refresh |
| `esc` | back | :back |
| `?` | help | :help |

### File Content

View file contents from containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :up |
| `down, j` | scroll down | :down |
| `G` | go to end | :go-to-end |
| `g` | go to start | :go-to-start |
| `esc` | back | :back |
| `?` | help | :help |

### Inspect View

View detailed container/image/network/volume information

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :up |
| `down, j` | scroll down | :down |
| `pgup` | page up | :page-up |
| `pgdown,  ` | page down | :page-down |
| `G` | go to end | :go-to-end |
| `g` | go to start | :go-to-start |
| `/` | search | :search |
| `n` | next match | :next-search-result |
| `N` | prev match | :prev-search-result |
| `esc` | back | :back |
| `?` | help | :help |

### Help View

View help information

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :up |
| `down, j` | scroll down | :down |
| `esc` | back | :back |

### Docker in Docker

View containers inside dind containers

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | move up | :up |
| `down, j` | move down | :down |
| `enter` | view logs | :log |
| `esc` | back | :back |
| `?` | help | :help |
| `x` | show actions | :show-actions |
| `f` | browse files | :file-browse |
| `!` | exec /bin/sh | :shell |
| `i` | inspect | :inspect |
| `r` | refresh | :refresh |
| `a` | toggle all | :toggle-all |
| `K` | kill | :kill |
| `S` | stop | :stop |
| `U` | start | :start |
| `R` | restart | :restart |
| `P` | pause/unpause | :pause |
| `D` | delete | :delete |
| `t` | top | :top |

### Command Execution

View command execution output

| Key | Description | Command |
|-----|-------------|----------|
| `up, k` | scroll up | :up |
| `down, j` | scroll down | :down |
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
