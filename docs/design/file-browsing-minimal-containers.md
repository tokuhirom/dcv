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

**Implementation:**
```go
type BinaryInjectorStrategy struct {
    client   *DockerClient
    injector *BinaryInjector
}

func (s *BinaryInjectorStrategy) ListDirectory(ctx context.Context, container, path string) ([]FileInfo, error) {
    busybox, err := s.injector.InjectBusybox(ctx, container)
    if err != nil {
        return nil, err
    }
    output, err := s.client.Exec(container, []string{busybox, "ls", "-la", path})
    return parseLS(output), nil
}
```

### Solution 2: Docker Export/Archive

Use Docker's export functionality to get the entire filesystem as a tar stream.

**Pros:**
- Always works, regardless of container contents
- No modification to container
- Can browse entire filesystem
- Read-only and safe

**Cons:**
- Slower for large containers
- Requires parsing tar stream
- May use significant bandwidth/memory for large containers

**Implementation:**
```go
type DockerExportStrategy struct {
    client *DockerClient
}

func (s *DockerExportStrategy) ListDirectory(ctx context.Context, container, path string) ([]FileInfo, error) {
    reader, err := s.client.ContainerExport(ctx, container)
    // Parse tar headers for the specific path
    return parseTarHeaders(reader, path), nil
}
```

### Solution 3: Shell Built-ins Fallback

Use shell built-ins when sh/bash exists but ls/cat don't.

**Pros:**
- No additional tools needed
- Works with minimal shells
- Fast and simple

**Cons:**
- Limited functionality
- Not all containers have shells
- Output parsing is fragile

**Implementation:**
```go
type ShellGlobStrategy struct {
    client *DockerClient
}

func (s *ShellGlobStrategy) ListDirectory(ctx context.Context, container, path string) ([]FileInfo, error) {
    // Use shell globbing
    cmd := fmt.Sprintf("cd %s && for f in * .[^.]*; do echo \"$f\"; done", path)
    output, err := s.client.Exec(container, []string{"sh", "-c", cmd})
    return parseGlobOutput(output), nil
}
```

### Solution 4: Sidecar Container (Not Recommended)

Start a debug container with shared namespaces/volumes.

**Pros:**
- Full debugging capabilities
- Target container untouched
- Rich toolset available

**Cons:**
- Complex lifecycle management
- Requires container creation permissions
- Only works with volumes, not container layer files
- Resource overhead
- Cleanup complexity

**Why not recommended:** Too complex for simple file browsing needs.

## Recommended Implementation Strategy

### Multi-Strategy Approach with Fallbacks

```go
type FileBrowserMultiStrategy struct {
    strategies []FileBrowserStrategy
}

func NewFileBrowserMultiStrategy(client *DockerClient) *FileBrowserMultiStrategy {
    return &FileBrowserMultiStrategy{
        strategies: []FileBrowserStrategy{
            &NativeToolsStrategy{client},     // Try native first (fastest)
            &ShellGlobStrategy{client},       // Try shell if available
            &BinaryInjectionStrategy{client}, // Inject tools if needed
            &DockerExportStrategy{client},    // Last resort
        },
    }
}

func (m *FileBrowserMultiStrategy) Browse(ctx context.Context, container, path string) ([]FileInfo, error) {
    for _, strategy := range m.strategies {
        if strategy.IsAvailable(ctx, container) {
            return strategy.ListDirectory(ctx, container, path)
        }
    }
    return nil, fmt.Errorf("no file browsing strategy available")
}
```

### Strategy Selection Logic

1. **Check native tools** - Fastest if available
2. **Check shell availability** - Use built-ins if possible
3. **Inject binary** - If allowed by configuration
4. **Use docker export** - Always works but slower

### Caching Layer

```go
type CachedFileBrowser struct {
    strategy FileBrowserStrategy
    cache    map[string]*CacheEntry
}

type CacheEntry struct {
    files     []FileInfo
    timestamp time.Time
}
```

## Configuration

Add to dcv configuration file:

```toml
[file_browser]
# Strategy selection: "auto", "native", "inject", "export"
strategy = "auto"

# Allow binary injection
allow_injection = true

# Allow docker export (may be slow)
allow_export = true

# Cleanup injected binaries on exit
cleanup_on_exit = true

# Cache directory listings (seconds)
cache_ttl = 30

# Custom busybox binary path
busybox_path = "~/.dcv/bin/busybox"
```

## User Experience

### First Time (No Tools Available)
```
[Enter] Browse files
> Detecting available tools...
> No native tools found
> Injecting helper binary... 
> Successfully injected /tmp/.dcv-busybox
> Listing directory /etc/...
```

### Subsequent Access (Cached)
```
[Enter] Browse files
> Using cached helper at /tmp/.dcv-busybox
> Listing directory /etc/...
```

### Fallback to Export
```
[Enter] Browse files
> No tools available and injection disabled
> Using docker export (this may be slow)...
> Listing directory /etc/...
```

## Security Considerations

1. **Binary Injection**
   - Only inject into /tmp (usually writable)
   - Use unique names to avoid conflicts
   - Verify binary checksum before injection
   - Cleanup on exit

2. **Permissions**
   - Respect user configuration
   - Don't inject if explicitly disabled
   - Log all injection activities

3. **Resource Usage**
   - Cache strategies per container
   - Limit export strategy for large containers
   - Implement timeouts

## Performance Optimizations

1. **Strategy Caching**
   - Remember which strategy works per container
   - Cache for container lifetime

2. **Directory Listing Cache**
   - Cache listings with TTL
   - Invalidate on container changes

3. **Binary Injection Reuse**
   - Check if binary exists before re-injecting
   - Share binary across multiple operations

## Implementation Phases

### Phase 1: Basic Multi-Strategy (MVP)
- Native tools strategy
- Docker export strategy
- Basic strategy selection

### Phase 2: Binary Injection
- BusyBox injection
- Cleanup mechanism
- Caching layer

### Phase 3: Advanced Features
- Shell built-ins strategy
- Configuration options
- Performance optimizations

### Phase 4: Future Enhancements
- Custom minimal file browser binary
- Sidecar container support for debug mode
- File search capabilities

## Testing Strategy

1. **Unit Tests**
   - Test each strategy independently
   - Mock Docker client responses
   - Test fallback logic

2. **Integration Tests**
   - Test with real containers:
     - Alpine (has tools)
     - Distroless (no tools)
     - Scratch (nothing)
   - Test cleanup mechanisms

3. **Performance Tests**
   - Measure strategy selection time
   - Compare performance of each strategy
   - Test caching effectiveness

## Alternatives Considered

### Docker CP for Listing
**Problem:** `docker cp` requires knowing paths beforehand, can't discover files.

### Direct Filesystem Access
**Problem:** Requires root access to Docker host, not portable.

### Permanent Container Modification
**Problem:** Changes container state, not acceptable for production.

## Conclusion

The multi-strategy approach with binary injection as the primary fallback provides the best balance of:
- **Compatibility** - Works with any container
- **Performance** - Fast after initial setup
- **Simplicity** - Transparent to users
- **Safety** - Temporary and reversible

This design ensures dcv's file browser works reliably across all container types while maintaining good performance and user experience.

## References

- [BusyBox Static Binaries](https://busybox.net/downloads/binaries/)
- [Docker SDK ContainerExport](https://pkg.go.dev/github.com/docker/docker/client#Client.ContainerExport)
- [Distroless Containers](https://github.com/GoogleContainerTools/distroless)
- [OpenTelemetry Collector Dockerfile](https://github.com/open-telemetry/opentelemetry-collector/blob/main/cmd/otelcol/Dockerfile)