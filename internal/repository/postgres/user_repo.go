package postgres

import (
	"database/sql"
	"fmt"
	"pull-request-reviewer-assignment-service/internal/models"
)

// предоставляет методы для работы с данными пользователей в базе данных
type UserRepository struct {
	db *sql.DB
}

// создает и возвращает новый экземпляр UserRepository
// принимает: подключение к базе данных для инициализации репозитория
// возвращает: указатель на созданный UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// сохраняет нового пользователя в базе данных
// принимает: указатель на объект User с данными для создания
// возвращает: ошибку в случае неудачного выполнения запроса к базе данных
func (r *UserRepository) CreateUser(user *models.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (user_id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
		user.UserID, user.Username, user.TeamName, user.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// возвращает данные пользователя по его идентификатору из базы данных
// принимает: строку с идентификатором пользователя для поиска
// возвращает: указатель на объект User с данными или ошибку если пользователь не найден
func (r *UserRepository) GetUser(userID string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// обновляет данные существующего пользователя в базе данных
// принимает: указатель на объект User с обновленными данными
// возвращает: ошибку в случае если пользователь не найден или произошла ошибка обновления
func (r *UserRepository) UpdateUser(user *models.User) error {
	result, err := r.db.Exec(
		"UPDATE users SET username = $1, team_name = $2, is_active = $3 WHERE user_id = $4",
		user.Username, user.TeamName, user.IsActive, user.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// возвращает список активных пользователей указанной команды
// принимает: строку с названием команды для поиска активных пользователей
// возвращает: слайс указателей на объекты User или ошибку выполнения запроса
func (r *UserRepository) GetActiveUsersByTeam(teamName string) ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = true 
		ORDER BY user_id
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// проверяет наличие пользователя с указанным идентификатором в базе данных
// принимает: строку с идентификатором пользователя для проверки существования
// возвращает: булево значение и ошибку, где true означает что пользователь существует
func (r *UserRepository) UserExists(userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)
	`, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}
