package models

import "time"

// представляет стандартизированный формат ответа с ошибкой API
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// содержит детали ошибки с кодом и сообщением
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// описывает структуру команды с названием и списком участников
type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

// представляет участника команды с информацией о активности
type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// описывает структуру пользователя системы
type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// содержит полную информацию о Pull Request
type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

// содержит сокращенную информацию о Pull Request
type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

// запрос на массовую деактивацию
type BulkDeactivateRequest struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids"`
}

// ответ массовой деактивации
type BulkDeactivateResponse struct {
	DeactivatedUsers []string       `json:"deactivated_users"`
	ReassignedPRs    []ReassignedPR `json:"reassigned_prs"`
	TotalProcessed   int            `json:"total_processed"`
	ReassignedCount  int            `json:"reassigned_count"`
}

// информация о переназначенных PR
type ReassignedPR struct {
	PRID         string   `json:"pr_id"`
	OldReviewers []string `json:"old_reviewers"`
	NewReviewers []string `json:"new_reviewers"`
}
