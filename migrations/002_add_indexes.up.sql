-- Индексы для оптимизации запросов

-- Для поиска активных пользователей в команде (логика назначения ревьюверов)
CREATE INDEX idx_users_team_active ON users(team_name, is_active);

-- Для поиска пользователей по активности
CREATE INDEX idx_users_active ON users(is_active);

-- Для поиска PR по автору
CREATE INDEX idx_pr_author ON pull_requests(author_id);

-- Для поиска PR по статусу
CREATE INDEX idx_pr_status ON pull_requests(status);

-- Для поиска назначений ревьюверов
CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);

-- Для сортировки PR по дате создания
CREATE INDEX idx_pr_created_at ON pull_requests(created_at);

-- Для поиска пользователей по команде
CREATE INDEX idx_users_team_name ON users(team_name);

-- Частичный индекс для открытых PR (самые частые запросы)
CREATE INDEX idx_pr_open_status ON pull_requests(status) WHERE status = 'OPEN';

-- Составной индекс для поиска активных пользователей в команде (исключая автора)
CREATE INDEX idx_users_team_active_composite ON users(team_name, is_active, user_id) WHERE is_active = true;