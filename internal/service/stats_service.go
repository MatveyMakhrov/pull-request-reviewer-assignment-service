package service

import (
	"pull-request-reviewer-assignment-service/internal/models"
	"pull-request-reviewer-assignment-service/internal/repository"
)

// предоставляет логику для работы со статистикой назначений
type StatsService struct {
	repo repository.StatsRepository
}

// создает и возвращает новый экземпляр StatsService
// принимает: репозиторий статистики для внедрения зависимости
// возвращает: указатель на созданный StatsService
func NewStatsService(repo repository.StatsRepository) *StatsService {
	return &StatsService{
		repo: repo,
	}
}

// возвращает агрегированную статистику по всем назначениям на код-ревью
// принимает: не принимает параметров, использует данные из репозитория статистики
// возвращает: указатель на StatsResponse с полной статистикой или ошибку получения данных
func (s *StatsService) GetReviewStats() (*models.StatsResponse, error) {
	userStats, err := s.repo.GetUserAssignmentStats()
	if err != nil {
		return nil, err
	}

	prStats, err := s.repo.GetPRAssignmentStats()
	if err != nil {
		return nil, err
	}

	totalAssignments := int64(0)
	for _, stat := range userStats {
		totalAssignments += stat.AssignmentCount
	}

	var topReviewers []models.UserAssignmentStats
	if len(userStats) > 5 {
		topReviewers = userStats[:5]
	} else {
		topReviewers = userStats
	}

	return &models.StatsResponse{
		TotalAssignments:  totalAssignments,
		AssignmentsByUser: userStats,
		AssignmentsByPR:   prStats,
		TopReviewers:      topReviewers,
	}, nil
}
