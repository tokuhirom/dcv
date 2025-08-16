# Design: File Browsing in Minimal Containers

## Problem Statement

Many modern containers (especially microservices like OpenTelemetry Collector) use minimal or distroless base images that lack basic Unix tools like `ls` and `cat`. This makes it impossible to browse files using dcv's current file browser implementation, which relies on executing these commands inside the container.

### Examples of Affected Containers:
- OpenTelemetry Collector (scratch/distroless)
- Google distroless images
- Scratch-based Go applications
- Minimal security-focused containers

## Current Implementation Limitations

The current file browser implementation in dcv:
1. Executes `ls -la` to list directories
2. Executes `cat` to read file contents
3. Fails completely when these tools don't exist

## Proposed Solutions

### Solution 1: Binary Injection (Recommended)

Inject a static binary (like BusyBox) into the container temporarily.

**Pros:**
- Universal solution - works on any Linux container
- Fast after initial injection
- Can be cached per container

**Cons:**
- Requires write access to container filesystem (usually /tmp)
- Needs cleanup mechanism
- Adds ~1MB to container temporarily
- May trigger security alerts in some environments

### Solution 2: Docker Export/Archive

Use Docker's export functionality to get the entire filesystem as a tar stream.

**Pros:**
- Always works, regardless of container contents
- No modification to container
- Can browse entire filesystem
- Read-only and safe

**Cons:**
- **Very Slow**
- May use significant bandwidth/memory for large containers

### Solution 3: Sidecar Container

Start a debug container with shared namespaces/volumes.

**Pros:**
- Full debugging capabilities
- Target container untouched
- Rich toolset available

**Cons:**
- Complex lifecycle management
- Only works with volumes, not container layer files
- Resource overhead
- Cleanup complexity

## References

- [BusyBox Static Binaries](https://busybox.net/downloads/binaries/)
- [Docker SDK ContainerExport](https://pkg.go.dev/github.com/docker/docker/client#Client.ContainerExport)
- [Distroless Containers](https://github.com/GoogleContainerTools/distroless)
