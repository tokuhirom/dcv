# Helper Binary Embedding Strategy

## Overview

The dcv-helper binaries need to be embedded into the main dcv binary so they can be injected into containers at runtime without requiring external downloads or dependencies.

## Build and Embedding Process

### 1. Two-Stage Build Process

```makefile
# Makefile
.PHONY: build-helpers
build-helpers:
	@echo "Building helper binaries..."
	@mkdir -p internal/docker/static-binaries
	
	# Build for amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -ldflags="-s -w" \
		-o internal/docker/static-binaries/dcv-helper-amd64 \
		cmd/dcv-helper/main.go
	
	# Build for arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -ldflags="-s -w" \
		-o internal/docker/static-binaries/dcv-helper-arm64 \
		cmd/dcv-helper/main.go
	
	# Build for 386
	CGO_ENABLED=0 GOOS=linux GOARCH=386 \
		go build -ldflags="-s -w" \
		-o internal/docker/static-binaries/dcv-helper-386 \
		cmd/dcv-helper/main.go
	
	@echo "Helper binaries built successfully"
	@ls -lh internal/docker/static-binaries/

.PHONY: build
build: build-helpers
	@echo "Building dcv with embedded helpers..."
	go build -o dcv
	@echo "Build complete!"

.PHONY: clean-helpers
clean-helpers:
	rm -f internal/docker/static-binaries/dcv-helper-*
```

### 2. Embedding in Go Code

```go
// internal/docker/embedded_helpers.go
package docker

import (
	_ "embed"
	"fmt"
	"runtime"
)

// Embed the helper binaries at compile time
// These files must exist before 'go build' is run

//go:embed static-binaries/dcv-helper-amd64
var helperBinaryAMD64 []byte

//go:embed static-binaries/dcv-helper-arm64
var helperBinaryARM64 []byte

//go:embed static-binaries/dcv-helper-386
var helperBinary386 []byte

// GetHelperBinary returns the appropriate helper binary for the current architecture
func GetHelperBinary(arch string) ([]byte, error) {
	// If arch not specified, use runtime arch
	if arch == "" {
		arch = runtime.GOARCH
	}
	
	switch arch {
	case "amd64", "x86_64":
		return helperBinaryAMD64, nil
	case "arm64", "aarch64":
		return helperBinaryARM64, nil
	case "386", "i386":
		return helperBinary386, nil
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", arch)
	}
}

// GetHelperBinarySize returns the size of embedded binary
func GetHelperBinarySize(arch string) int {
	binary, err := GetHelperBinary(arch)
	if err != nil {
		return 0
	}
	return len(binary)
}
```

### 3. CI/CD Integration

```yaml
# .github/workflows/build.yml
name: Build

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build-helpers:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        arch: [amd64, arm64, 386]
    
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build helper for ${{ matrix.arch }}
      run: |
        CGO_ENABLED=0 GOOS=linux GOARCH=${{ matrix.arch }} \
          go build -ldflags="-s -w" \
          -o internal/docker/static-binaries/dcv-helper-${{ matrix.arch }} \
          cmd/dcv-helper/main.go
    
    - name: Upload helper binary
      uses: actions/upload-artifact@v3
      with:
        name: helper-${{ matrix.arch }}
        path: internal/docker/static-binaries/dcv-helper-${{ matrix.arch }}
  
  build-dcv:
    needs: build-helpers
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Download helper binaries
      uses: actions/download-artifact@v3
      with:
        path: internal/docker/static-binaries/
    
    - name: Move helpers to correct location
      run: |
        mv internal/docker/static-binaries/helper-*/* internal/docker/static-binaries/
        rmdir internal/docker/static-binaries/helper-*
        ls -la internal/docker/static-binaries/
    
    - name: Build dcv with embedded helpers
      run: go build -o dcv
    
    - name: Test embedded helpers
      run: |
        # Verify helpers are embedded
        go test -v ./internal/docker -run TestEmbeddedHelpers
```

## Directory Structure

```
dcv/
├── cmd/
│   ├── dcv-helper/
│   │   └── main.go              # Helper binary source
│   └── generate-screenshots/
│       └── main.go
├── internal/
│   └── docker/
│       ├── static-binaries/      # Generated helper binaries (git-ignored)
│       │   ├── .gitkeep
│       │   ├── dcv-helper-amd64
│       │   ├── dcv-helper-arm64
│       │   └── dcv-helper-386
│       ├── embedded_helpers.go   # Embedding logic
│       └── binary_injector.go    # Injection logic
├── .gitignore                    # Ignore generated binaries
└── Makefile                      # Build automation
```

### .gitignore Addition
```gitignore
# Generated helper binaries
internal/docker/static-binaries/dcv-helper-*

# Keep the directory
!internal/docker/static-binaries/.gitkeep
```

## Size Impact Analysis

### Binary Size Comparison

| Component | Size | Notes |
|-----------|------|-------|
| dcv (without helpers) | ~15 MB | Base application |
| dcv-helper-amd64 | ~120 KB | Most common |
| dcv-helper-arm64 | ~120 KB | ARM servers |
| dcv-helper-386 | ~110 KB | Legacy systems |
| **Total embedded** | ~350 KB | All architectures |
| **dcv (with helpers)** | ~15.3 MB | Only 2% increase |

### Runtime Memory Impact

The embedded binaries are only loaded into memory when needed:

```go
func (bi *BinaryInjector) InjectHelper(ctx context.Context, containerID string) (string, error) {
    // Only load the binary for the target architecture
    arch := bi.detectContainerArch(ctx, containerID)
    binaryData, err := GetHelperBinary(arch)
    // binaryData is now in memory (120KB)
    
    // After injection, can be garbage collected
    return bi.copyToContainer(ctx, containerID, binaryData, "/tmp/.dcv-helper")
}
```

## Development Workflow

### For Contributors

1. **First-time setup:**
   ```bash
   git clone https://github.com/tokuhirom/dcv
   cd dcv
   make build-helpers  # Build helper binaries
   make build          # Build dcv with embedded helpers
   ```

2. **Modifying helper:**
   ```bash
   # Edit cmd/dcv-helper/main.go
   vim cmd/dcv-helper/main.go
   
   # Rebuild helpers and dcv
   make clean-helpers
   make build
   
   # Test
   ./dcv
   ```

3. **Testing without rebuilding dcv:**
   ```bash
   # Build and test helper standalone
   go build -o test-helper cmd/dcv-helper/main.go
   ./test-helper ls
   ./test-helper cat README.md
   ```

## Alternative Approaches Considered

### 1. Download at Runtime
**Pros:** Smaller dcv binary
**Cons:** Requires internet, security concerns, slower first run
**Decision:** Rejected - reliability issues

### 2. Build Helper Inside Container
**Pros:** No embedding needed
**Cons:** Requires Go compiler in container
**Decision:** Rejected - impossible for minimal containers

### 3. Ship Separate Binary Files
**Pros:** Smaller main binary
**Cons:** Complex distribution, path management
**Decision:** Rejected - poor user experience

### 4. Use go:generate
**Pros:** Automatic generation
**Cons:** Requires binaries before build
**Decision:** Partially adopted - use Makefile instead

## Implementation Checklist

- [ ] Create cmd/dcv-helper/main.go
- [ ] Update Makefile with build-helpers target
- [ ] Create embedded_helpers.go
- [ ] Update .gitignore
- [ ] Update CI/CD workflow
- [ ] Add tests for embedding
- [ ] Update build documentation
- [ ] Test on all architectures
- [ ] Verify binary size impact
- [ ] Add version checking

## Testing Strategy

### Unit Test for Embedding
```go
func TestEmbeddedHelpers(t *testing.T) {
    architectures := []string{"amd64", "arm64", "386"}
    
    for _, arch := range architectures {
        t.Run(arch, func(t *testing.T) {
            binary, err := GetHelperBinary(arch)
            require.NoError(t, err)
            require.NotEmpty(t, binary)
            require.Greater(t, len(binary), 50000) // At least 50KB
            require.Less(t, len(binary), 200000)   // Less than 200KB
            
            // Verify it's a valid ELF binary
            require.True(t, bytes.HasPrefix(binary, []byte{0x7f, 'E', 'L', 'F'}))
        })
    }
}
```

## Conclusion

Embedding the helper binaries directly into dcv provides:
- **Zero dependencies** at runtime
- **No network required** for operation  
- **Fast injection** (binary already in memory)
- **Secure** (no external downloads)
- **Simple distribution** (single binary)

The ~350KB size increase (2%) is negligible compared to the benefits of having a fully self-contained tool that works with any container.