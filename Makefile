.PHONY: build run test clean dev loadtest quicktest logs db-connect stats health help

# Основные команды - запускают только app и postgres
build:
	docker-compose build

run:
	docker-compose up

dev:
	docker-compose up -d --build

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

# Локальная разработка (без Docker)
local-build:
	go build -o server ./cmd/server

local-run: local-build
	DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=password DB_NAME=pr_reviewer ./server

# Помощь
help:
	@echo "Available commands:"
	@echo "  build     - Собрать Docker образы"
	@echo "  run       - Запустить основной проект (app + postgres)"
	@echo "  dev       - Запустить с пересборкой"
	@echo "  loadtest  - Нагрузочное тестирование (отдельная команда)"
	@echo "  quicktest - Быстрое нагрузочное тестирование через k6"
	@echo "  test      - Запустить unit-тесты"
	@echo "  logs      - Просмотр логов приложения"
	@echo "  db-connect- Подключиться к БД"
	@echo "  stats     - Показать статистику"
	@echo "  health    - Проверить health check"
	@echo "  clean     - Остановить и очистить"
	@echo "  help      - Показать эту справку"