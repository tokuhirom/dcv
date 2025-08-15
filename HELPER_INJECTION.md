# Helper Binary Injection

## Overview
The dcv-helper binary enables file browsing in containers without standard tools (ls, cat, etc.). It's now injected using `docker cp` instead of Docker API for better compatibility.

## Commands
- `:inject-helper` - Manually inject helper binary (available in file browser and dind views)

## Usage

### Regular Containers
In file browser view, type `:inject-helper` to inject the helper when automatic injection fails.

### Docker-in-Docker (dind)
In dind process list view, type `:inject-helper` to inject helper into selected nested container. This enables file browsing in containers running inside dind.

## Implementation
- Uses `docker cp` to copy helper binary to `/tmp/.dcv-helper`
- For dind: injects to host container first, then copies to nested container
- Automatic fallback when native commands fail
- Manual injection for explicit control