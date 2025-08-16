.PHONY: all test fmt dev-deps lint clean screenshots screenshots-deps clean-helpers

all: build-helpers
	go build -o dcv

# Run tests
test:
	go test -v ./...

# Format code with golangci-lint
fmt:
	@echo "Running code formatter..."
	golangci-lint fmt .

# Run golangci-lint
lint:
	@which golangci-lint > /dev/null || $(MAKE) dev-deps
	golangci-lint run

# Clean build artifacts
clean: clean-helpers
	rm -f dcv

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@which golangci-lint > /dev/null || { \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest; \
		echo "golangci-lint installed successfully"; \
	}
	@which lefthook > /dev/null || go install github.com/evilmartians/lefthook@latest
	@lefthook install
	@echo "Development dependencies installed successfully"

# Install screenshot generation dependencies
screenshots-deps:
	@echo "Installing screenshot dependencies..."
	@go get github.com/pavelpatrin/go-ansi-to-image@latest
	@echo "Screenshot dependencies installed successfully"

# Generate screenshots for documentation
screenshots: screenshots-deps
	@echo "Generating screenshots..."
	@go run -tags screenshots cmd/generate-screenshots/main.go
	@echo "Screenshots generated successfully!"

# Define helper binary targets
HELPER_BINARIES = internal/docker/static-binaries/dcv-helper-amd64 \
                  internal/docker/static-binaries/dcv-helper-arm64 \
                  internal/docker/static-binaries/dcv-helper-arm

# Build helper binaries for embedding
build-helpers: $(HELPER_BINARIES)

internal/docker/static-binaries/dcv-helper-amd64: cmd/dcv-helper/main.go
	@echo "  Building for linux/amd64 (x86_64)..."
	@mkdir -p internal/docker/static-binaries
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -ldflags="-s -w" -trimpath \
		-o $@ \
		cmd/dcv-helper/main.go
	@strip $@ 2>/dev/null || true

internal/docker/static-binaries/dcv-helper-arm64: cmd/dcv-helper/main.go
	@echo "  Building for linux/arm64 (aarch64)..."
	@mkdir -p internal/docker/static-binaries
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -ldflags="-s -w" -trimpath \
		-o $@ \
		cmd/dcv-helper/main.go
	@strip $@ 2>/dev/null || true

internal/docker/static-binaries/dcv-helper-arm: cmd/dcv-helper/main.go
	@echo "  Building for linux/arm (armv7/armhf)..."
	@mkdir -p internal/docker/static-binaries
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
		go build -ldflags="-s -w" -trimpath \
		-o $@ \
		cmd/dcv-helper/main.go
	@strip $@ 2>/dev/null || true

# Clean helper binaries
clean-helpers:
	@echo "Cleaning helper binaries..."
	@rm -f internal/docker/static-binaries/dcv-helper-*
