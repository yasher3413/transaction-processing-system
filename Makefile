.PHONY: up down test lint fmt seed e2e clean migrate-up migrate-down

# Ensure Docker is in PATH
export PATH := /Applications/Docker.app/Contents/Resources/bin:$(PATH)

# Start all services
up:
	cd infra && docker compose up --build -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services are up! API: http://localhost:8080, Grafana: http://localhost:3000"

# Stop all services
down:
	cd infra && docker compose down

# Run all tests
test:
	cd tests && go test -v ./...

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...

# Seed sample data
seed:
	@echo "Seeding sample accounts..."
	@curl -X POST http://localhost:8080/v1/accounts \
		-H "Content-Type: application/json" \
		-H "X-API-Key: demo-api-key-12345" \
		-d '{"currency": "USD"}' || true
	@curl -X POST http://localhost:8080/v1/accounts \
		-H "Content-Type: application/json" \
		-H "X-API-Key: demo-api-key-12345" \
		-d '{"currency": "EUR"}' || true

# Run E2E tests
e2e:
	cd tests/e2e && go test -v -timeout 5m

# Clean up
clean:
	cd infra && docker compose down -v
	rm -rf bin/ dist/

# Run migrations up
migrate-up:
	@echo "Running migrations..."
	@docker compose -f infra/docker-compose.yml exec -T postgres psql -U postgres -d transactions < infra/migrations/001_initial_schema.sql || true

# Run migrations down (reset)
migrate-down:
	@echo "Resetting database..."
	@docker compose -f infra/docker-compose.yml exec -T postgres psql -U postgres -d transactions -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" || true

