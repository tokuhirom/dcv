.PHONY: all up down logs test-dind clean staticcheck test

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
	@which staticcheck > /dev/null || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...

test:
	go test -v ./...
