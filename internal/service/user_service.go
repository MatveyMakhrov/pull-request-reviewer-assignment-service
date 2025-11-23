package service

import (
	"fmt"
	"log"
	"pull-request-reviewer-assignment-service/internal/models"
	"pull-request-reviewer-assignment-service/internal/repository"
	"time"
)

// предоставляет логику для работы с пользователями и их активностью
type UserService struct {
	userRepo   repository.UserRepository
	prRepo     repository.PRRepository
	teamRepo   repository.TeamRepository
	reviewRepo repository.ReviewRepository
}

// создает и возвращает новый экземпляр UserService
// принимает: репозитории пользователей, PR, команд и ревью для внедрения зависимостей
// возвращает: указатель на созданный UserService
func NewUserService(userRepo repository.UserRepository, prRepo repository.PRRepository,
	teamRepo repository.TeamRepository, reviewRepo repository.ReviewRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		prRepo:     prRepo,
		teamRepo:   teamRepo,
		reviewRepo: reviewRepo,
	}
}

// изменяет статус активности пользователя и сохраняет изменения в базе данных
// принимает: идентификатор пользователя и булево значение для установки активности
// возвращает: обновленный объект User или ошибку если пользователь не найден
func (s *UserService) SetUserActive(userID string, isActive bool) (*models.User, error) {
	log.Printf("Setting user activity: %s -> %t", userID, isActive)

	// получаем пользователя
	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		log.Printf("User not found: %s, error: %v", userID, err)
		return nil, NewServiceError("NOT_FOUND", "user not found")
	}

	// обновляем активность
	user.IsActive = isActive

	// сохраняем изменения
	if err := s.userRepo.UpdateUser(user); err != nil {
		log.Printf("Failed to update user: %s, error: %v", userID, err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("User activity updated: %s -> %t", userID, isActive)
	return user, nil
}

// возвращает данные пользователя по его идентификатору
// принимает: строку с идентификатором пользователя для поиска в репозитории
// возвращает: указатель на объект User или ошибку если пользователь не найден
func (s *UserService) GetUser(userID string) (*models.User, error) {
	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		return nil, NewServiceError("NOT_FOUND", "user not found")
	}
	return user, nil
}

// возвращает список Pull Request назначенных пользователю на ревью если пользователь активен
// принимает: строку с идентификатором пользователя для поиска назначенных PR
// возвращает: слайс сокращенных объектов PullRequestShort или ошибку если пользователь не найден
func (s *UserService) GetUserReviewPRs(userID string) ([]*models.PullRequestShort, error) {
	log.Printf("Getting PRs for user review: %s", userID)

	// проверяем существование пользователя и его активность
	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		log.Printf("User not found: %s, error: %v", userID, err)
		return nil, NewServiceError("NOT_FOUND", "user not found")
	}

	// проверяем что пользователь активен
	if !user.IsActive {
		log.Printf("User %s is inactive, returning empty PR list", userID)
		return []*models.PullRequestShort{}, nil
	}

	// получаем PR из репозитория
	prs, err := s.prRepo.GetPRsByReviewer(userID)
	if err != nil {
		log.Printf("Failed to get PRs for user: %s, error: %v", userID, err)
		return nil, fmt.Errorf("failed to get user PRs: %w", err)
	}

	log.Printf("Found %d PRs for user: %s", len(prs), userID)
	return prs, nil
}

// массово деактивирует пользователей команды и переназначает их открытые PR на других ревьюверов
// принимает: название команды и список идентификаторов пользователей для деактивации
// возвращает: объект BulkDeactivateResponse со статистикой операции или ошибку выполнения
func (s *UserService) BulkDeactivateUsers(teamName string, userIDs []string) (*models.BulkDeactivateResponse, error) {
	startTime := time.Now()
	log.Printf("Starting bulk deactivation for team %s, users: %v", teamName, userIDs)

	teamExists, err := s.teamRepo.TeamExists(teamName)
	if err != nil {
		return nil, NewServiceError("INTERNAL_ERROR", err.Error())
	}
	if !teamExists {
		return nil, NewServiceError("NOT_FOUND", "team not found")
	}

	deactivatedUsers := make([]string, 0)
	for _, userID := range userIDs {
		user, err := s.userRepo.GetUser(userID)
		if err != nil {
			continue
		}

		if user.TeamName != teamName {
			continue
		}

		user.IsActive = false
		if err := s.userRepo.UpdateUser(user); err != nil {
			return nil, NewServiceError("INTERNAL_ERROR", err.Error())
		}

		deactivatedUsers = append(deactivatedUsers, userID)
		log.Printf("User deactivated: %s", userID)
	}

	reassignedPRs := make([]models.ReassignedPR, 0)

	for _, userID := range deactivatedUsers {
		openPRs, err := s.getOpenPRsWithReviewer(userID)
		if err != nil {
			log.Printf("Failed to get open PRs for user %s: %v", userID, err)
			continue
		}

		log.Printf("User %s has %d open PRs for reassignment", userID, len(openPRs))

		for _, pr := range openPRs {
			reassignedPR, err := s.reassignReviewerInPR(pr.PullRequestID, userID, teamName)
			if err != nil {
				log.Printf("Failed to reassign PR %s: %v", pr.PullRequestID, err)
				continue
			}

			if reassignedPR != nil {
				reassignedPRs = append(reassignedPRs, *reassignedPR)
				log.Printf("PR %s reassigned: %s -> %s", pr.PullRequestID, userID, reassignedPR.NewReviewers)
			}
		}
	}

	// проверяем время выполнения
	executionTime := time.Since(startTime)
	log.Printf("Bulk deactivation completed in %v", executionTime)

	if executionTime > 100*time.Millisecond {
		log.Printf("Bulk deactivation took %v (target < 100ms)", executionTime)
	}

	return &models.BulkDeactivateResponse{
		DeactivatedUsers: deactivatedUsers,
		ReassignedPRs:    reassignedPRs,
		TotalProcessed:   len(deactivatedUsers),
		ReassignedCount:  len(reassignedPRs),
	}, nil
}

// возвращает список открытых Pull Request где пользователь назначен ревьювером
// принимает: идентификатор пользователя для поиска назначенных открытых PR
// возвращает: слайс полных объектов PullRequest или ошибку выполнения запроса
func (s *UserService) getOpenPRsWithReviewer(userID string) ([]*models.PullRequest, error) {
	// Получаем все PR пользователя
	prShorts, err := s.prRepo.GetPRsByReviewer(userID)
	if err != nil {
		return nil, err
	}

	// Фильтруем только открытые PR
	var openPRs []*models.PullRequest
	for _, prShort := range prShorts {
		if prShort.Status == "OPEN" {
			fullPR, err := s.prRepo.GetPR(prShort.PullRequestID)
			if err != nil {
				continue
			}
			openPRs = append(openPRs, fullPR)
		}
	}

	return openPRs, nil
}

// переназначает одного ревьювера на другого активного пользователя из той же команды в Pull Request
// принимает: идентификатор PR, идентификатор старого ревьювера и название команды для поиска замены
// возвращает: объект ReassignedPR с информацией о переназначении или ошибку выполнения операции
func (s *UserService) reassignReviewerInPR(prID, oldReviewerID, teamName string) (*models.ReassignedPR, error) {
	log.Printf("Reassigning reviewer in PR %s: %s -> ?", prID, oldReviewerID)

	// получаем текущих ревьюверов
	currentReviewers, err := s.reviewRepo.GetAssignedReviewers(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigned reviewers: %w", err)
	}

	// проверяем, что старый ревьювер действительно назначен
	if !contains(currentReviewers, oldReviewerID) {
		return nil, fmt.Errorf("reviewer %s not assigned to PR %s", oldReviewerID, prID)
	}

	// получаем информацию о PR
	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	// находим активных пользователей команды для замены
	availableUsers, err := s.userRepo.GetActiveUsersByTeam(teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	// фильтруем доступных кандидатов
	var candidates []string
	for _, user := range availableUsers {
		if user.UserID != pr.AuthorID &&
			user.UserID != oldReviewerID &&
			!contains(currentReviewers, user.UserID) {
			candidates = append(candidates, user.UserID)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available candidates for replacement in team %s", teamName)
	}

	// выбираем первого кандидата
	newReviewerID := candidates[0]

	// выполняем замену
	if err := s.reviewRepo.ReplaceReviewer(prID, oldReviewerID, newReviewerID); err != nil {
		return nil, fmt.Errorf("failed to replace reviewer: %w", err)
	}

	// обновляем список ревьюверов для ответа
	newReviewers := make([]string, len(currentReviewers))
	copy(newReviewers, currentReviewers)
	for i, reviewer := range newReviewers {
		if reviewer == oldReviewerID {
			newReviewers[i] = newReviewerID
			break
		}
	}

	log.Printf("Successfully reassigned PR %s: %s -> %s", prID, oldReviewerID, newReviewerID)

	return &models.ReassignedPR{
		PRID:         prID,
		OldReviewers: currentReviewers,
		NewReviewers: newReviewers,
	}, nil
}

// вспомогательная функция для проверки наличия элемента в слайсе
// принимает: слайс строк и строку для поиска в этом слайсе
// возвращает: true если строка найдена в слайсе, иначе false
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
