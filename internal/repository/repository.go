package repository

import "pull-request-reviewer-assignment-service/internal/models"

// интерфейс для работы с командами
type TeamRepository interface {
	CreateTeam(team *models.Team) error
	GetTeam(teamName string) (*models.Team, error)
	TeamExists(teamName string) (bool, error)
}

// интерфейс для работы с пользователями
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUser(userID string) (*models.User, error)
	UpdateUser(user *models.User) error
	GetActiveUsersByTeam(teamName string) ([]*models.User, error)
	UserExists(userID string) (bool, error)
}

// интерфейс для работы с pull requests
type PRRepository interface {
	CreatePR(pr *models.PullRequest) error
	GetPR(prID string) (*models.PullRequest, error)
	UpdatePR(pr *models.PullRequest) error
	PRExists(prID string) (bool, error)
	GetPRsByReviewer(userID string) ([]*models.PullRequestShort, error)
}

// интерфейс для работы с ревьюверами
type ReviewRepository interface {
	AssignReviewers(prID string, reviewerIDs []string) error
	GetAssignedReviewers(prID string) ([]string, error)
	ReplaceReviewer(prID, oldReviewerID, newReviewerID string) error
	IsReviewerAssigned(prID, userID string) (bool, error)
}

// интерфейс для работы со статистикой
type StatsRepository interface {
	GetUserAssignmentStats() ([]models.UserAssignmentStats, error)
	GetPRAssignmentStats() ([]models.PRAssignmentStats, error)
}
