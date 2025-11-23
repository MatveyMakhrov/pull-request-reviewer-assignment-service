package service

import (
	"fmt"
	"log"
	"pull-request-reviewer-assignment-service/internal/models"
	"pull-request-reviewer-assignment-service/internal/repository"
)

// предоставляет логику для работы с командами и их участниками
type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

// создает и возвращает новый экземпляр TeamService
// принимает: репозитории команд и пользователей для внедрения зависимостей
// возвращает: указатель на созданный TeamService
func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// создает новую команду и всех её участников после валидации данных
// принимает: указатель на объект Team с данными команды и списком участников
// возвращает: ошибку если команда уже существует или данные участников невалидны
func (s *TeamService) CreateTeam(team *models.Team) error {
	log.Printf("Creating team: %s with %d members", team.TeamName, len(team.Members))

	// проверяем существование команды
	exists, err := s.teamRepo.TeamExists(team.TeamName)
	if err != nil {
		log.Printf("Failed to check team existence: %v", err)
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		log.Printf("Team already exists: %s", team.TeamName)
		return NewServiceError("TEAM_EXISTS", "team_name already exists")
	}

	// валидация участников
	for i, member := range team.Members {
		if member.UserID == "" {
			return NewServiceError("INVALID_REQUEST", fmt.Sprintf("user_id is required for member %d", i))
		}
		if member.Username == "" {
			return NewServiceError("INVALID_REQUEST", fmt.Sprintf("username is required for member %d", i))
		}
	}

	log.Printf("Team validation passed, creating team: %s", team.TeamName)

	// создаем команду
	if err := s.teamRepo.CreateTeam(team); err != nil {
		log.Printf("Failed to create team: %v", err)
		return fmt.Errorf("failed to create team: %w", err)
	}

	log.Printf("Team created successfully: %s", team.TeamName)
	return nil
}

// возвращает полную информацию о команде включая список всех участников
// принимает: строку с названием команды для поиска в репозитории
// возвращает: указатель на объект Team с данными или ошибку если команда не найдена
func (s *TeamService) GetTeam(teamName string) (*models.Team, error) {
	log.Printf("Getting team: %s", teamName)

	team, err := s.teamRepo.GetTeam(teamName)
	if err != nil {
		log.Printf("Team not found: %s, error: %v", teamName, err)
		return nil, NewServiceError("NOT_FOUND", "team not found")
	}

	log.Printf("Team found: %s with %d members", teamName, len(team.Members))
	return team, nil
}
