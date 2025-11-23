package postgres

import (
	"database/sql"
	"fmt"
)

// предоставляет методы для работы с данными о ревью в базе данных
type ReviewRepository struct {
	db *sql.DB
}

// создает и возвращает новый экземпляр ReviewRepository
// принимает: подключение к базе данных для инициализации репозитория
// возвращает: указатель на созданный ReviewRepository
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// назначает нескольких ревьюверов на указанный Pull Request
// принимает: идентификатор PR и слайс идентификаторов ревьюверов для назначения
// возвращает: ошибку в случае неудачного выполнения транзакции назначения
func (r *ReviewRepository) AssignReviewers(prID string, reviewerIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, reviewerID := range reviewerIDs {
		_, err = tx.Exec(`
			INSERT INTO pr_reviewers (pull_request_id, reviewer_id) 
			VALUES ($1, $2)
		`, prID, reviewerID)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	return tx.Commit()
}

// возвращает список ревьюверов назначенных на указанный Pull Request
// принимает: строку с идентификатором Pull Request для поиска назначенных ревьюверов
// возвращает: слайс строк с идентификаторами ревьюверов или ошибку выполнения запроса
func (r *ReviewRepository) GetAssignedReviewers(prID string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT reviewer_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1 
		ORDER BY assigned_at
	`, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to query assigned reviewers: %w", err)
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

// заменяет одного ревьювера на другого в указанном Pull Request
// принимает: идентификатор PR, идентификатор старого ревьювера и идентификатор нового ревьювера
// возвращает: ошибку если старый ревьювер не был назначен или произошла ошибка замены
func (r *ReviewRepository) ReplaceReviewer(prID, oldReviewerID, newReviewerID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// удаляем старого ревьювера
	result, err := tx.Exec(`
		DELETE FROM pr_reviewers 
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`, prID, oldReviewerID)
	if err != nil {
		return fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reviewer not assigned to this PR")
	}

	// добавляем нового ревьювера
	_, err = tx.Exec(`
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id) 
		VALUES ($1, $2)
	`, prID, newReviewerID)
	if err != nil {
		return fmt.Errorf("failed to assign new reviewer: %w", err)
	}

	return tx.Commit()
}

// проверяет назначен ли указанный пользователь ревьювером на Pull Request
// принимает: идентификатор PR и идентификатор пользователя для проверки назначения
// возвращает: булево значение и ошибку, где true означает что пользователь назначен ревьювером
func (r *ReviewRepository) IsReviewerAssigned(prID, userID string) (bool, error) {
	var assigned bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pr_reviewers 
			WHERE pull_request_id = $1 AND reviewer_id = $2
		)
	`, prID, userID).Scan(&assigned)
	if err != nil {
		return false, fmt.Errorf("failed to check reviewer assignment: %w", err)
	}
	return assigned, nil
}
