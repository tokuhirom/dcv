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

# Run staticcheck
staticcheck:
	@which staticcheck > /dev/null || $(MAKE) dev-deps
	staticcheck ./...

# Run lint (alias for staticcheck)
lint: staticcheck

test:
	go test -v ./...

# Format code with goimports (fallback to go fmt if goimports fails)
fmt:
	@echo "Running code formatter..."
	@if command -v goimports >/dev/null 2>&1 && goimports -w . 2>/dev/null; then \
		echo "Formatted with goimports"; \
	else \
		echo "goimports not available, using go fmt"; \
		go fmt ./...; \
	fi

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Development dependencies installed successfully"
