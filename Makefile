.PHONY: all up down logs test-dind clean staticcheck test fmt dev-deps lint

all:
	go build -o dcv

# Start all services
up:
	docker compose up -d
	@echo "Waiting for services to start..."
	@sleep 5
	@echo "Starting containers inside dind..."
	@docker compose exec -T dind sh /startup.sh || true

# Stop all services
down:
	docker compose down -v

# Show logs
logs:
	docker compose logs -f

# Setup dind containers
test-dind:
	docker compose exec -T dind sh /startup.sh

# Clean everything
clean:
	docker compose down -v
	rm -f dcv

# Run golangci-lint
lint:
	@which golangci-lint > /dev/null || $(MAKE) install-golangci-lint
	golangci-lint run

# Install golangci-lint
install-golangci-lint:
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
	@echo "golangci-lint installed successfully"

# Run staticcheck (kept for backward compatibility)
staticcheck: lint

test:
	go test -v ./...

# Format code with goimports (fallback to go fmt if goimports fails)
fmt:
	@echo "Running code formatter..."
	golangci-lint fmt .

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@which golangci-lint > /dev/null || $(MAKE) install-golangci-lint
	@which lefthook > /dev/null || go install github.com/evilmartians/lefthook@latest
	@lefthook install
	@echo "Development dependencies installed successfully"
