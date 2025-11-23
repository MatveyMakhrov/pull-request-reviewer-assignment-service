package postgres

import (
	"database/sql"
	"fmt"
	"pull-request-reviewer-assignment-service/internal/models"
)

// предоставляет методы для работы с данными Pull Request в базе данных
type PRRepository struct {
	db *sql.DB
}

// создает и возвращает новый экземпляр PRRepository
// принимает: подключение к базе данных для инициализации репозитория
// возвращает: указатель на созданный PRRepository
func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{db: db}
}

// сохраняет новый Pull Request в базе данных
// принимает: указатель на объект PullRequest с данными для создания
// возвращает: ошибку в случае неудачного выполнения запроса к базе данных
func (r *PRRepository) CreatePR(pr *models.PullRequest) error {
	_, err := r.db.Exec(`
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	return nil
}

// возвращает полную информацию о Pull Request по его идентификатору
// принимает: строку с идентификатором Pull Request для поиска в базе данных
// возвращает: указатель на объект PullRequest с данными или ошибку если PR не найден
func (r *PRRepository) GetPR(prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	var mergedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests 
		WHERE pull_request_id = $1
	`, prID).Scan(
		&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status,
		&pr.CreatedAt, &mergedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pull request not found")
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	// получаем назначенных ревьюверов
	reviewers, err := r.getPRReviewers(prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

// обновляет данные существующего Pull Request в базе данных
// принимает: указатель на объект PullRequest с обновленными данными
// возвращает: ошибку в случае если PR не найден или произошла ошибка обновления
func (r *PRRepository) UpdatePR(pr *models.PullRequest) error {
	var mergedAt interface{}
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	} else {
		mergedAt = nil
	}

	result, err := r.db.Exec(`
		UPDATE pull_requests 
		SET pull_request_name = $1, author_id = $2, status = $3, merged_at = $4 
		WHERE pull_request_id = $5
	`, pr.PullRequestName, pr.AuthorID, pr.Status, mergedAt, pr.PullRequestID)
	if err != nil {
		return fmt.Errorf("failed to update pull request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pull request not found")
	}

	return nil
}

// проверяет наличие Pull Request с указанным идентификатором в базе данных
// принимает: строку с идентификатором Pull Request для проверки существования
// возвращает: булево значение и ошибку, где true означает что PR существует
func (r *PRRepository) PRExists(prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)
	`, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}
	return exists, nil
}

// возвращает список идентификаторов ревьюверов назначенных на Pull Request
// принимает: строку с идентификатором Pull Request для поиска назначенных ревьюверов
// возвращает: слайс строк с идентификаторами ревьюверов или ошибку выполнения запроса
func (r *PRRepository) getPRReviewers(prID string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT reviewer_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1 
		ORDER BY assigned_at
	`, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to query PR reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviewers: %w", err)
	}

	return reviewers, nil
}

// возвращает список идентификаторов ревьюверов назначенных на указанный Pull Request
// принимает: строку с идентификатором Pull Request для получения списка ревьюверов
// возвращает: слайс строк с идентификаторами ревьюверов или ошибку выполнения запроса
func (r *PRRepository) GetPRReviewers(prID string) ([]string, error) {
	return r.getPRReviewers(prID)
}

// возвращает список Pull Request назначенных пользователю на ревью
// принимает: строку с идентификатором пользователя для поиска назначенных PR
// возвращает: слайс сокращенных объектов PullRequestShort или ошибку выполнения запроса
func (r *PRRepository) GetPRsByReviewer(userID string) ([]*models.PullRequestShort, error) {
	rows, err := r.db.Query(`
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}

	return prs, nil
}
