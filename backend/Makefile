DC = docker compose -f ../docker-compose.yml

APP = guiltmachine-backend
BIN = ./bin/$(APP)
MAIN = ./cmd/api
DB_URL = postgres://guilt:guiltpass@localhost:5432/guiltmachine?sslmode=disable

.PHONY: run build db-up db-down db-wait migrate-up app-up stop clean logs help redis-up redis-down redis-wait services logs-all test test-up test-down test-wait test-db test-db-drop

run: db-up redis-up db-wait migrate-up build app-up

db-up:
	$(DC) up -d postgres

db-down:
	$(DC) stop postgres || true

db-wait:
	@echo "waiting for postgres..."
	@until $(DC) exec -T postgres pg_isready -U guilt -d guiltmachine >/dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "postgres ready"

redis-up:
	$(DC) up -d redis

redis-down:
	$(DC) stop redis || true

redis-wait:
	@echo "waiting for redis..."
	@until $(DC) exec -T redis redis-cli ping >/dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "redis ready"

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

services:
	$(DC) up -d postgres redis

stop:
	$(DC) down || true
	pkill -f $(BIN) || true

clean:
	rm -rf ./bin
	$(DC) down -v

logs:
	$(DC) logs -f postgres

logs-all:
	$(DC) logs -f

test:
	$(MAKE) -C test test

test-up:
	$(MAKE) -C test test-up

test-down:
	$(MAKE) -C test test-down

test-wait:
	$(MAKE) -C test test-wait

test-db:
	$(MAKE) -C test test-db

test-db-drop:
	@echo "  make test        - Run full test suite with infrastructure"
	@echo "  make test-up     - Start test infrastructure"
	@echo "  make test-down   - Stop test infrastructure"
	@echo "  make test-wait   - Wait for test postgres"
	@echo "  make test-db     - Create test database and migrations"
	@echo "  make test-db-drop - Drop test database"
	$(MAKE) -C test test-db-drop

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make run         - Start postgres + redis, wait for postgres, run migrations, build and run the app"
	@echo "  make db-up       - Start postgres"
	@echo "  make redis-up    - Start redis"
	@echo "  make db-wait     - Wait for postgres to be ready"
	@echo "  make redis-wait  - Wait for redis to be ready"
	@echo "  make migrate-up  - Run database migrations"
	@echo "  make build       - Build the app binary"
	@echo "  make app-up      - Run the built app"
	@echo "  make services    - Start postgres + redis"
	@echo "  make stop        - Stop services and kill app process"
	@echo "  make clean       - Remove build artifacts and stop all containers"
	@echo "  make logs        - Show postgres logs"
	@echo "  make logs-all    - Show logs for all services"
	@echo "  make envoy-up    - Start envoy proxy"
	@echo "  make envoy-logs  - Show envoy logs"
	@echo "  make backend-up  - Start backend container"
	@echo "  make full-stack  - Start all services (postgres, redis, backend, envoy)"

# Envoy targets
.PHONY: envoy-up envoy-logs envoy-down backend-up backend-logs full-stack

envoy-up:
	$(DC) up -d envoy

envoy-logs:
	$(DC) logs -f envoy

envoy-down:
	$(DC) stop envoy || true

backend-up:
	$(DC) up -d backend

backend-logs:
	$(DC) logs -f backend

backend-down:
	$(DC) stop backend || true

full-stack:
	$(DC) up -d postgres redis
	@$(MAKE) db-wait
	@$(MAKE) redis-wait
	@$(MAKE) migrate-up
	$(DC) up -d backend envoy
	@echo "Full stack running:"
	@echo "  - Postgres: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - Backend gRPC: localhost:9090"
	@echo "  - Envoy (gRPC-Web): localhost:8080"
	@echo "  - Envoy Admin: localhost:8001"

