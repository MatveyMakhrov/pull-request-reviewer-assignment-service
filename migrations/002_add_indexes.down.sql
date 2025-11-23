-- Удаление индексов
DROP INDEX IF EXISTS idx_users_team_active;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_pr_author;
DROP INDEX IF EXISTS idx_pr_status;
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP INDEX IF EXISTS idx_pr_created_at;
DROP INDEX IF EXISTS idx_users_team_name;
DROP INDEX IF EXISTS idx_pr_open_status;
DROP INDEX IF EXISTS idx_users_team_active_composite;