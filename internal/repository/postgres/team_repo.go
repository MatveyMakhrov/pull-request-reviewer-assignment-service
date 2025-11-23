package postgres

import (
	"database/sql"
	"fmt"
	"pull-request-reviewer-assignment-service/internal/models"
)

// предоставляет методы для работы с данными команд в базе данных
type TeamRepository struct {
	db *sql.DB
}

// создает и возвращает новый экземпляр TeamRepository
// принимает: подключение к базе данных для инициализации репозитория
// возвращает: указатель на созданный TeamRepository
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// CreateTeam создает команду и ее участников в транзакции
func (r *TeamRepository) CreateTeam(team *models.Team) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// вставляем команду
	_, err = tx.Exec("INSERT INTO teams (team_name) VALUES ($1)", team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to insert team: %w", err)
	}

	// вставляем пользователей
	for _, member := range team.Members {
		_, err = tx.Exec(
			"INSERT INTO users (user_id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
			member.UserID, member.Username, team.TeamName, member.IsActive,
		)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %w", member.UserID, err)
		}
	}

	return tx.Commit()
}

// возвращает команду с участниками
// принимает: указатель на объект Team с данными команды и списком участников
// возвращает: ошибку в случае неудачного выполнения транзакции создания
func (r *TeamRepository) GetTeam(teamName string) (*models.Team, error) {
	// Получаем основную информацию о команде
	var team models.Team
	team.TeamName = teamName

	// Получаем участников команды
	rows, err := r.db.Query(`
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1 
		ORDER BY user_id
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	team.Members = members
	return &team, nil
}

// проверяет наличие команды с указанным названием в базе данных
// принимает: строку с названием команды для проверки существования
// возвращает: булево значение и ошибку, где true означает что команда существует
func (r *TeamRepository) TeamExists(teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)
	`, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}
	return exists, nil
}
