DC = docker compose -f ../docker-compose.yml

APP = guiltmachine-backend
BIN = ./bin/$(APP)
MAIN = ./cmd/api
DB_URL = postgres://guilt:guiltpass@localhost:5432/guiltmachine?sslmode=disable

.PHONY: run build db-up db-down db-wait migrate-up app-up stop clean logs help

run: db-up db-wait migrate-up build app-up

db-up:
	$(DC) up -d postgres

db-down:
	$(DC) down

db-wait:
	@echo "waiting for postgres..."
	@until $(DC) exec -T postgres pg_isready -U guilt -d guiltmachine >/dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "postgres ready"

migrate-up:
	@echo "running migrations..."
	@which migrate > /dev/null 2>&1 || (echo "migrate-cli not found. Installing..."; go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path ./migrations -database "$(DB_URL)" up
	@echo "migrations completed"

app-up:
	$(BIN)

build:
	@mkdir -p ./bin
	go build -o $(BIN) $(MAIN)

stop:
	$(DC) down || true
	pkill -f $(BIN) || true

clean:
	rm -rf ./bin
	$(DC) down -v

logs:
	$(DC) logs -f postgres

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make run      - Start postgres, wait for readiness, run migrations, build and run the app"
	@echo "  make db-up    - Start postgres container"
	@echo "  make db-down  - Stop postgres container"
	@echo "  make db-wait  - Wait for postgres to be ready"
	@echo "  make migrate-up - Run database migrations"
	@echo "  make build    - Build the app binary"
	@echo "  make app-up   - Run the built app"
	@echo "  make stop     - Stop all services and kill app process"
	@echo "  make clean    - Remove build artifacts and stop all containers"
	@echo "  make logs     - Show postgres container logs"
