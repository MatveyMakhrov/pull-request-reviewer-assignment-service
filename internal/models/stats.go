package models

// ответ статистики
type StatsResponse struct {
	TotalAssignments  int64                 `json:"total_assignments"`
	AssignmentsByUser []UserAssignmentStats `json:"assignments_by_user"`
	AssignmentsByPR   []PRAssignmentStats   `json:"assignments_by_pr"`
	TopReviewers      []UserAssignmentStats `json:"top_reviewers"`
}

// представляет статистику назначений для конкретного пользователя
type UserAssignmentStats struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	AssignmentCount int64  `json:"assignment_count"`
}

// представляет статистику назначений для конкретного Pull Request
type PRAssignmentStats struct {
	PRID            string `json:"pr_id"`
	PRName          string `json:"pr_name"`
	AssignmentCount int64  `json:"assignment_count"`
}
