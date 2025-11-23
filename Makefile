.PHONY: build run clean dev loadtest logs db-connect stats health help
.PHONY: lint lint-fix lint-handlers lint-services lint-models lint-repositories setup-linter

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
	@echo "  help      - Показать эту справку"