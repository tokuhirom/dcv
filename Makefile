.PHONY: all test fmt dev-deps lint clean

all:
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
clean:
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
