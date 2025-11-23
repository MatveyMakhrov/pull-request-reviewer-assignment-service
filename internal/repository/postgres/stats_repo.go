package postgres

import (
	"context"
	"database/sql"
	"pull-request-reviewer-assignment-service/internal/models"
)

// предоставляет методы для работы со статистикой в базе данных
type StatsRepository struct {
	db *sql.DB
}

// создает и возвращает новый экземпляр StatsRepository
// принимает: подключение к базе данных для инициализации репозитория
// возвращает: указатель на созданный StatsRepository
func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// возвращает статистику назначений на код-ревью по активным пользователям
// принимает: ничего, использует контекст по умолчанию для выполнения запроса
// возвращает: слайс структур UserAssignmentStats с количеством назначений или ошибку
func (r *StatsRepository) GetUserAssignmentStats() ([]models.UserAssignmentStats, error) {
	query := `
        SELECT u.user_id, u.username, COUNT(pr.reviewer_id) as assignment_count
        FROM users u
        LEFT JOIN pr_reviewers pr ON u.user_id = pr.reviewer_id
        WHERE u.is_active = true
        GROUP BY u.user_id, u.username
        ORDER BY assignment_count DESC
    `

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.UserAssignmentStats
	for rows.Next() {
		var stat models.UserAssignmentStats
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.AssignmentCount); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// возвращает статистику назначений ревьюверов по всем Pull Request
// принимает: ничего, использует контекст по умолчанию для выполнения запроса
// возвращает: слайс структур PRAssignmentStats с количеством назначений на каждый PR или ошибку
func (r *StatsRepository) GetPRAssignmentStats() ([]models.PRAssignmentStats, error) {
	query := `
        SELECT p.pull_request_id, p.pull_request_name, COUNT(pr.reviewer_id) as assignment_count
        FROM pull_requests p
        LEFT JOIN pr_reviewers pr ON p.pull_request_id = pr.pull_request_id
        GROUP BY p.pull_request_id, p.pull_request_name
        ORDER BY assignment_count DESC
    `

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PRAssignmentStats
	for rows.Next() {
		var stat models.PRAssignmentStats
		if err := rows.Scan(&stat.PRID, &stat.PRName, &stat.AssignmentCount); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}
