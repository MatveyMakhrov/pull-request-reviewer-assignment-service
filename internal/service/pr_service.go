package service

import (
	"fmt"
	"log"
	"math/rand"
	"pull-request-reviewer-assignment-service/internal/models"
	"pull-request-reviewer-assignment-service/internal/repository"
	"time"
)

// предоставляет логику для работы с Pull Request
type PRService struct {
	prRepo      repository.PRRepository
	reviewRepo  repository.ReviewRepository
	userRepo    repository.UserRepository
	teamService *TeamService
}

// создает и возвращает новый экземпляр PRService с внедренными зависимостями
// принимает: репозитории PR, ревью, пользователей и сервис команд для инициализации
// возвращает: указатель на созданный PRService с инициализированным генератором случайных чисел
func NewPRService(prRepo repository.PRRepository, reviewRepo repository.ReviewRepository, userRepo repository.UserRepository, teamService *TeamService) *PRService {
	// инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	return &PRService{
		prRepo:      prRepo,
		reviewRepo:  reviewRepo,
		userRepo:    userRepo,
		teamService: teamService,
	}
}

// создает новый Pull Request и автоматически назначает ревьюверов из команды автора
// принимает: идентификатор PR, название PR и идентификатор автора для создания
// возвращает: указатель на созданный PullRequest или ошибку валидации/назначения
func (s *PRService) CreatePR(prID, prName, authorID string) (*models.PullRequest, error) {
	log.Printf("Creating PR: %s by author: %s", prID, authorID)

	// проверяем существование PR
	exists, err := s.prRepo.PRExists(prID)
	if err != nil {
		log.Printf("Failed to check PR existence: %s, error: %v", prID, err)
		return nil, fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		log.Printf("PR already exists: %s", prID)
		return nil, NewServiceError("PR_EXISTS", "PR id already exists")
	}

	// проверяем существование автора
	author, err := s.userRepo.GetUser(authorID)
	if err != nil {
		log.Printf("Author not found: %s, error: %v", authorID, err)
		return nil, NewServiceError("NOT_FOUND", "author not found")
	}

	// проверяем что автор активен
	if !author.IsActive {
		log.Printf("Author is not active: %s", authorID)
		return nil, NewServiceError("INVALID_REQUEST", "author is not active")
	}

	// назначаем ревьюверов
	reviewerIDs, err := s.assignReviewers(authorID, author.TeamName)
	if err != nil {
		log.Printf("Failed to assign reviewers: %v", err)
		return nil, fmt.Errorf("failed to assign reviewers: %w", err)
	}

	log.Printf("Assigned reviewers for PR %s: %v", prID, reviewerIDs)

	// создаем PR
	pr := &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewerIDs,
		CreatedAt:         time.Now(),
	}

	if err := s.prRepo.CreatePR(pr); err != nil {
		log.Printf("Failed to create PR: %s, error: %v", prID, err)
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	// назначаем ревьюверов в отдельной таблице
	if len(reviewerIDs) > 0 {
		if err := s.reviewRepo.AssignReviewers(prID, reviewerIDs); err != nil {
			log.Printf("Failed to assign reviewers to PR: %s, error: %v", prID, err)
		}
	}

	log.Printf("PR created successfully: %s with %d reviewers", prID, len(reviewerIDs))
	return pr, nil
}

// помечает Pull Request как MERGED (идемпотентная операция)
// принимает: идентификатор Pull Request для выполнения операции мержа
// возвращает: обновленный объект PullRequest или ошибку если PR не найден или не может быть мержен
func (s *PRService) MergePR(prID string) (*models.PullRequest, error) {
	log.Printf("Merging PR: %s", prID)

	// получаем PR
	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		log.Printf("PR not found: %s, error: %v", prID, err)
		return nil, NewServiceError("NOT_FOUND", "PR not found")
	}

	// проверяем текущий статус
	if pr.Status == "MERGED" {
		log.Printf("PR already merged: %s, returning current state", prID)
		// Идемпотентность - возвращаем текущее состояние без ошибки
		return pr, nil
	}

	// проверяем что PR открыт
	if pr.Status != "OPEN" {
		log.Printf("PR is not open: %s, status: %s", prID, pr.Status)
		return nil, NewServiceError("INVALID_REQUEST", "cannot merge PR that is not open")
	}

	// обновляем статус и время мержа
	now := time.Now()
	pr.Status = "MERGED"
	pr.MergedAt = &now

	// сохраняем изменения
	if err := s.prRepo.UpdatePR(pr); err != nil {
		log.Printf("Failed to merge PR: %s, error: %v", prID, err)
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	log.Printf("PR merged successfully: %s at %v", prID, now)
	return pr, nil
}

// assignReviewers назначает до 2 активных ревьюверов из команды автора
func (s *PRService) assignReviewers(authorID, teamName string) ([]string, error) {
	log.Printf("Assigning reviewers for author: %s from team: %s", authorID, teamName)

	// получаем активных пользователей команды
	activeUsers, err := s.userRepo.GetActiveUsersByTeam(teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	log.Printf("Found %d active users in team %s", len(activeUsers), teamName)

	// фильтруем автора и выбираем случайных ревьюверов
	var candidateUserIDs []string
	for _, user := range activeUsers {
		if user.UserID != authorID {
			candidateUserIDs = append(candidateUserIDs, user.UserID)
		}
	}

	log.Printf("Available reviewers (excluding author): %v", candidateUserIDs)

	if len(candidateUserIDs) == 0 {
		log.Printf("No available reviewers in team %s", teamName)
		return []string{}, nil
	}

	// выбираем до 2 случайных ревьюверов
	reviewerCount := min(2, len(candidateUserIDs))
	selectedReviewers := make([]string, 0, reviewerCount)

	// перемешиваем кандидатов
	shuffledCandidates := make([]string, len(candidateUserIDs))
	copy(shuffledCandidates, candidateUserIDs)
	rand.Shuffle(len(shuffledCandidates), func(i, j int) {
		shuffledCandidates[i], shuffledCandidates[j] = shuffledCandidates[j], shuffledCandidates[i]
	})

	// выбираем первых reviewerCount кандидатов
	for i := 0; i < reviewerCount; i++ {
		selectedReviewers = append(selectedReviewers, shuffledCandidates[i])
	}

	log.Printf("Selected %d reviewers: %v", len(selectedReviewers), selectedReviewers)
	return selectedReviewers, nil
}

// возвращает минимальное значение из двух целых чисел
// Принимает: два целых числа a и b для сравнения
// возвращает: наименьшее из двух переданных чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// переназначает ревьювера на другого активного пользователя из той же команды
// принимает: идентификатор PR и идентификатор старого ревьювера для замены
// возвращает: обновленный PR, идентификатор нового ревьювера или ошибку валидации/замены
func (s *PRService) ReassignReviewer(prID, oldReviewerID string) (*models.PullRequest, string, error) {
	log.Printf("Reassigning reviewer: %s in PR: %s", oldReviewerID, prID)

	// получаем PR
	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		log.Printf("PR not found: %s, error: %v", prID, err)
		return nil, "", NewServiceError("NOT_FOUND", "PR not found")
	}

	// проверяем что PR не мержен
	if pr.Status == "MERGED" {
		log.Printf("Cannot reassign on merged PR: %s", prID)
		return nil, "", NewServiceError("PR_MERGED", "cannot reassign on merged PR")
	}

	// проверяем что старый ревьювер назначен на PR
	isAssigned, err := s.reviewRepo.IsReviewerAssigned(prID, oldReviewerID)
	if err != nil {
		log.Printf("Failed to check reviewer assignment: %s in PR: %s, error: %v", oldReviewerID, prID, err)
		return nil, "", fmt.Errorf("failed to check reviewer assignment: %w", err)
	}
	if !isAssigned {
		log.Printf("Reviewer not assigned: %s in PR: %s", oldReviewerID, prID)
		return nil, "", NewServiceError("NOT_ASSIGNED", "reviewer is not assigned to this PR")
	}

	// получаем информацию о старом ревьювере
	oldReviewer, err := s.userRepo.GetUser(oldReviewerID)
	if err != nil {
		log.Printf("Old reviewer not found: %s, error: %v", oldReviewerID, err)
		return nil, "", NewServiceError("NOT_FOUND", "old reviewer not found")
	}

	// проверяем что старый ревьювер активен
	if !oldReviewer.IsActive {
		log.Printf("Old reviewer is not active: %s", oldReviewerID)
		return nil, "", NewServiceError("INVALID_REQUEST", "old reviewer is not active")
	}

	// выбираем нового ревьювера из команды старого ревьювера
	newReviewerID, err := s.selectReplacementReviewer(oldReviewer.TeamName, prID, pr.AuthorID, oldReviewerID)
	if err != nil {
		log.Printf("Failed to select replacement reviewer: %v", err)
		return nil, "", err
	}

	// заменяем ревьювера
	if err := s.reviewRepo.ReplaceReviewer(prID, oldReviewerID, newReviewerID); err != nil {
		log.Printf("Failed to replace reviewer: %s -> %s in PR: %s, error: %v", oldReviewerID, newReviewerID, prID, err)
		return nil, "", fmt.Errorf("failed to replace reviewer: %w", err)
	}

	// обновляем список ревьюверов в объекте PR
	pr.AssignedReviewers = s.replaceInSlice(pr.AssignedReviewers, oldReviewerID, newReviewerID)

	log.Printf("Reviewer reassigned successfully: %s -> %s in PR: %s", oldReviewerID, newReviewerID, prID)
	return pr, newReviewerID, nil
}

// выбирает случайного активного пользователя из команды для замены ревьювера
// принимает: название команды, идентификаторы PR, автора и старого ревьювера для фильтрации кандидатов
// возвращает: идентификатор выбранного пользователя или ошибку если нет подходящих кандидатов
func (s *PRService) selectReplacementReviewer(teamName, prID, authorID, oldReviewerID string) (string, error) {
	log.Printf("Selecting replacement reviewer from team: %s", teamName)

	// получаем активных пользователей команды
	activeUsers, err := s.userRepo.GetActiveUsersByTeam(teamName)
	if err != nil {
		return "", fmt.Errorf("failed to get active users: %w", err)
	}

	log.Printf("Found %d active users in team %s", len(activeUsers), teamName)

	// фильтруем кандидатов
	var candidateUserIDs []string
	for _, user := range activeUsers {
		// исключаем автора, старого ревьювера и уже назначенных ревьюверов
		if user.UserID != authorID &&
			user.UserID != oldReviewerID &&
			!s.isReviewerAssignedToPR(prID, user.UserID) {
			candidateUserIDs = append(candidateUserIDs, user.UserID)
		}
	}

	log.Printf("Available replacement candidates: %v", candidateUserIDs)

	if len(candidateUserIDs) == 0 {
		log.Printf("No available replacement candidates in team %s", teamName)
		return "", NewServiceError("NO_CANDIDATE", "no active replacement candidate in team")
	}

	// выбираем случайного кандидата
	selectedReviewer := candidateUserIDs[rand.Intn(len(candidateUserIDs))]
	log.Printf("Selected replacement reviewer: %s", selectedReviewer)
	return selectedReviewer, nil
}

// проверяет назначен ли указанный пользователь ревьювером на Pull Request
// принимает: идентификатор PR и идентификатор пользователя для проверки назначения
// возвращает: булево значение true если пользователь назначен ревьювером на PR
func (s *PRService) isReviewerAssignedToPR(prID, userID string) bool {
	assigned, err := s.reviewRepo.IsReviewerAssigned(prID, userID)
	if err != nil {
		log.Printf("Failed to check if user %s is assigned to PR %s: %v", userID, prID, err)
		return false
	}
	return assigned
}

// заменяет все вхождения старого элемента на новый в слайсе строк
// принимает: исходный слайс, старую строку для замены и новую строку для вставки
// возвращает: новый слайс с выполненными заменами элементов
func (s *PRService) replaceInSlice(slice []string, old, new string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		if item == old {
			result[i] = new
		} else {
			result[i] = item
		}
	}
	return result
}
