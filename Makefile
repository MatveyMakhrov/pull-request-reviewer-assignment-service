.PHONY: build run clean dev loadtest logs db-connect stats health help
.PHONY: lint lint-fix lint-handlers lint-services lint-models lint-repositories setup-linter
.PHONY: test-e2e setup-e2e-deps run-e2e-tests stop-e2e

# ==================== DOCKER COMMANDS ====================
build:
	docker-compose build

run:
	docker-compose up

dev:
	docker-compose up --build

clean:
	docker-compose down -v
	docker system prune -f

loadtest:
	docker-compose --profile loadtest run --rm loadtest

logs:
	docker-compose logs -f app

db-connect:
	docker-compose exec postgres psql -U postgres -d pr_reviewer

stats:
	curl http://localhost:8080/stats/review-assignments

health:
	curl http://localhost:8080/health

local-build:
	go build -o server ./cmd/server

local-run: local-build
	DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=password DB_NAME=pr_reviewer ./server

# ==================== CODE QUALITY COMMANDS ====================

setup-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run ./... --fix

lint-handlers:
	golangci-lint run ./internal/handlers/...

lint-services:
	golangci-lint run ./internal/service/...

lint-models:
	golangci-lint run ./internal/models/...

lint-repositories:
	golangci-lint run ./internal/repository/...

lint-fast:
	golangci-lint run ./... --fast

fmt:
	go fmt ./...

imports:
	goimports -l -w .

style: fmt imports lint

pre-commit: lint-fix test

# ==================== E2E COMMANDS ====================

setup-e2e-deps:
	go get github.com/stretchr/testify/assert
	go get github.com/stretchr/testify/suite

run-e2e-env:
	docker-compose -f docker-compose.e2e.yml up -d
	@echo "Waiting for services to start..."
	@sleep 15

stop-e2e-env:
	docker-compose -f docker-compose.e2e.yml down -v

test-e2e: setup-e2e-deps run-e2e-env
	@echo "Running E2E tests..."
	go test -v -tags=e2e ./tests/e2e/... -timeout=5m
	$(MAKE) stop-e2e-env

test-e2e-fast: run-e2e-env
	go test -v -tags=e2e ./tests/e2e/... -timeout=3m
	$(MAKE) stop-e2e-env

e2e-env: run-e2e-env
	@echo "E2E environment is running:"
	@echo "App: http://localhost:8082"
	@echo "DB: localhost:5434"

help:
	@echo "Available commands:"
	@echo ""
	@echo "=== DOCKER COMMANDS ==="
	@echo "  build     - Собрать Docker образы"
	@echo "  run       - Запустить основной проект"
	@echo "  dev       - Запустить с пересборкой"
	@echo "  loadtest  - Нагрузочное тестирование"
	@echo "  logs      - Просмотр логов приложения"
	@echo "  db-connect- Подключиться к БД"
	@echo "  stats     - Показать статистику"
	@echo "  health    - Проверить health check"
	@echo "  clean     - Остановить и очистить"
	@echo ""
	@echo "=== CODE QUALITY COMMANDS ==="
	@echo "  setup-linter - Установить линтер"
	@echo "  lint        - Проверить весь код"
	@echo "  lint-fix    - Исправить автоматически исправимые проблемы"
	@echo "  lint-handlers - Линтинг только обработчиков HTTP"
	@echo "  lint-services - Линтинг только сервисов"
	@echo "  lint-models - Линтинг только моделей"
	@echo "  lint-repositories - Линтинг только репозиториев"
	@echo "  lint-fast   - Быстрый линтинг (только критические проверки)"
	@echo "  fmt         - Форматировать код"
	@echo "  imports     - Организовать импорты"
	@echo "  style       - Полная проверка стиля (fmt + imports + lint)"
	@echo "  pre-commit  - Запустить проверки перед коммитом"
	@echo ""
	@echo "=== LOCAL DEVELOPMENT ==="
	@echo "  local-build - Собрать приложение локально"
	@echo "  local-run   - Запустить приложение локально"
	@echo ""
	@echo "=== E2E TESTING COMMANDS ==="
	@echo "  setup-e2e-deps - Установить зависимости для E2E тестов"
	@echo "  test-e2e       - Запустить полные E2E тесты"
	@echo "  test-e2e-fast  - Быстрые E2E тесты (без пересборки)"
	@echo "  test-e2e-simple- Упрощенные E2E тесты"
	@echo "  e2e-env        - Запустить E2E окружение для дебага"
	@echo "  run-e2e-env    - Только запустить E2E окружение"
	@echo "  stop-e2e-env   - Остановить E2E окружение"
	@echo ""
	@echo "  help           - Показать эту справку"