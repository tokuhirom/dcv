# Implementation Plan: Enhanced File Browser

## Quick Summary

Enhance dcv's file browser to work with minimal/distroless containers that lack basic Unix tools.

## Current Status

- ❌ Fails on containers without `ls`/`cat`
- ❌ No fallback mechanisms
- ❌ Users cannot browse files in minimal containers

## Target Status

- ✅ Works with ANY container
- ✅ Automatic fallback strategies
- ✅ Transparent to users

## Implementation Steps

### Step 1: Create Strategy Interface (Week 1)

```go
type FileBrowserStrategy interface {
    Name() string
    IsAvailable(ctx context.Context, container string) bool
    ListDirectory(ctx context.Context, container, path string) ([]FileInfo, error)
    ReadFile(ctx context.Context, container, path string) (string, error)
}
```

**Files to modify:**
- `internal/docker/file_browser.go` - Add interface
- `internal/docker/file_browser_strategies.go` - Implement strategies

### Step 2: Implement Native Tools Strategy (Week 1)

Current implementation refactored as a strategy.

**Testing:**
- Alpine Linux container ✅
- Ubuntu container ✅

### Step 3: Implement Docker Export Strategy (Week 2)

Fallback that always works.

**Key code:**
```go
func (s *DockerExportStrategy) ListDirectory(...) {
    reader, _ := dockerClient.ContainerExport(ctx, containerID)
    tarReader := tar.NewReader(reader)
    // Parse headers matching path
}
```

**Testing:**
- Distroless container ✅
- Scratch container ✅

### Step 4: Implement Binary Injection Strategy (Week 3)

Most powerful fallback.

**Tasks:**
1. Download/embed BusyBox static binary
2. Implement injection mechanism
3. Add cleanup tracking
4. Test on various architectures

**Testing:**
- OpenTelemetry Collector ✅
- Custom minimal containers ✅

### Step 5: Add Configuration (Week 4)

```toml
[file_browser]
strategy = "auto"
allow_injection = true
allow_export = true
```

**Files to modify:**
- `internal/config/config.go`
- `dcv.toml.example`

### Step 6: Add UI Feedback (Week 4)

Show strategy being used:
```
📁 Browsing files (using binary injection)...
```

**Files to modify:**
- `internal/ui/view_file_browser.go`

## Testing Matrix

| Container Type | Native | Export | Injection | Expected |
|---------------|--------|---------|-----------|----------|
| Alpine | ✅ | - | - | Use Native |
| Ubuntu | ✅ | - | - | Use Native |
| Distroless | ❌ | ✅ | ✅ | Use Injection |
| Scratch | ❌ | ✅ | ✅ | Use Injection |
| OTel Collector | ❌ | ✅ | ✅ | Use Injection |
| Busybox | ✅ | - | - | Use Native |

## Success Metrics

1. **Compatibility**: Works with 100% of Linux containers
2. **Performance**: < 1s to browse directories (after first access)
3. **User Experience**: No manual intervention required
4. **Reliability**: Automatic cleanup, no residual files

## Rollout Plan

### Phase 1: Internal Testing
- Test with common containers
- Gather performance metrics
- Fix edge cases

### Phase 2: Beta Release
- Add feature flag: `ENABLE_ENHANCED_FILE_BROWSER=true`
- Document in README
- Gather user feedback

### Phase 3: General Availability
- Enable by default
- Add to main documentation
- Create troubleshooting guide

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Binary injection blocked by security | Fall back to export strategy |
| Large containers slow with export | Add size warning, implement streaming |
| Cleanup fails leaving binaries | Track in persistent state, cleanup on next run |
| Architecture mismatch | Detect arch, download appropriate binary |

## Code Review Checklist

- [ ] All strategies implement interface
- [ ] Fallback logic tested
- [ ] Cleanup mechanisms in place
- [ ] Configuration documented
- [ ] Error messages helpful
- [ ] Performance acceptable
- [ ] Security considerations addressed

## Documentation Updates

1. **README.md** - Add feature description
2. **docs/keymap.md** - No changes needed
3. **CLAUDE.md** - Add design decision
4. **docs/troubleshooting.md** - Add file browser section

## Future Enhancements

1. **Custom lightweight binary** - Build our own minimal ls/cat binary
2. **Sidecar debug mode** - Full debugging container
3. **File search** - Find files across filesystem
4. **File edit** - Modify files (with confirmation)
5. **File download** - Export files to host

## Questions to Resolve

1. Should we embed BusyBox binary or download on demand?
   - **Decision:** Download on demand, cache locally

2. Default strategy when all are available?
   - **Decision:** Native > Injection > Export (speed priority)

3. How to handle Windows containers?
   - **Decision:** Out of scope for now

## Dependencies

- No new Go dependencies required
- BusyBox static binary (external, downloaded)
- Docker API version compatibility check

## Timeline

- **Week 1-2:** Core implementation
- **Week 3:** Testing and refinement  
- **Week 4:** Configuration and UI
- **Week 5:** Documentation and release

## Definition of Done

- [ ] All strategies implemented and tested
- [ ] Configuration options working
- [ ] Documentation updated
- [ ] Tests passing (>80% coverage)
- [ ] Code reviewed and approved
- [ ] Performance benchmarks met
- [ ] User feedback incorporated