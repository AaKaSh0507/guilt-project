.PHONY: test test-up test-down test-wait test-db test-db-drop

TEST_DC = docker compose -f docker/docker-compose.test.yml

test-up:
	$(TEST_DC) up -d

test-down:
	$(TEST_DC) down -v || true

test-wait:
	@echo "waiting for test postgres..."
	@until docker exec guilt_postgres_test pg_isready -U guilt >/dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "postgres ready"

test-db:
	@bash scripts/create_test_db.sh

test-db-drop:
	@bash scripts/drop_test_db.sh

test:
	$(MAKE) test-up
	$(MAKE) test-wait
	@export TEST_DB_URL=$$(bash scripts/create_test_db.sh | tail -1) && \
	export TEST_REDIS_URL=redis://localhost:6380 && \
	go test ../test/... ../internal/... -count=1
	$(MAKE) test-db-drop
	$(MAKE) test-down
