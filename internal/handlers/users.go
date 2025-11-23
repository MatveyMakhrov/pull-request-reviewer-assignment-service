package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"pull-request-reviewer-assignment-service/internal/models"
	"pull-request-reviewer-assignment-service/internal/service"
)

// обрабатывает HTTP запросы связанные с пользователями
type UserHandler struct {
	userService *service.UserService
}

// создает и возвращает новый экземпляр UserHandler
// принимает: сервис пользователей для внедрения зависимости
// возвращает: указатель на созданный UserHandler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// обрабатывает изменение активности пользователя
// принимает: HTTP запрос с JSON содержащим user_id и is_active
// возвращает: JSON с обновленными данными пользователя или ошибку
func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received POST /users/setIsActive request")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Invalid JSON: %v", err)
		writeError(w, "INVALID_REQUEST", "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed request: user_id=%s, is_active=%t", request.UserID, request.IsActive)

	// валидация
	if request.UserID == "" {
		log.Printf("Missing user_id")
		writeError(w, "INVALID_REQUEST", "user_id is required", http.StatusBadRequest)
		return
	}

	// изменяем активность пользователя через сервис
	log.Printf("Calling user service to update user: %s", request.UserID)
	user, err := h.userService.SetUserActive(request.UserID, request.IsActive)
	if err != nil {
		log.Printf("Service error: %v", err)
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Code {
			case "NOT_FOUND":
				writeError(w, "NOT_FOUND", serviceErr.Message, http.StatusNotFound)
			case "INVALID_REQUEST":
				writeError(w, "INVALID_REQUEST", serviceErr.Message, http.StatusBadRequest)
			default:
				writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("User activity updated successfully: %s -> %t", request.UserID, request.IsActive)
	response := map[string]interface{}{
		"user": user,
	}
	writeJSON(w, http.StatusOK, response)
}

// обрабатывает получение PR пользователя для ревью
// принимает: HTTP GET запрос с параметром user_id в URL
// возвращает: JSON со списком PR и идентификатором пользователя или ошибку
func (h *UserHandler) GetUserReviewPRs(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received GET /users/getReview request")

	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s", r.Method)
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		log.Printf("Missing user_id parameter")
		writeError(w, "INVALID_REQUEST", "user_id parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("Getting PRs for user: %s", userID)

	// получаем PR пользователя через сервис
	log.Printf("Calling user service to get PRs for user: %s", userID)
	prs, err := h.userService.GetUserReviewPRs(userID)
	if err != nil {
		log.Printf("Service error: %v", err)
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Code {
			case "NOT_FOUND":
				writeError(w, "NOT_FOUND", serviceErr.Message, http.StatusNotFound)
			case "INVALID_REQUEST":
				writeError(w, "INVALID_REQUEST", serviceErr.Message, http.StatusBadRequest)
			default:
				writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d PRs for user: %s", len(prs), userID)

	response := map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	}
	writeJSON(w, http.StatusOK, response)
}

// обрабатывает массовую деактивацию пользователей
// принимает: HTTP запрос с JSON содержащим team_name и список user_ids для деактивации
// возвращает: JSON со статистикой выполненной операции или ошибку валидации/выполнения
func (h *UserHandler) BulkDeactivate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received POST /users/bulk-deactivate request")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request models.BulkDeactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Invalid JSON: %v", err)
		writeError(w, "INVALID_REQUEST", "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed request: team=%s, users=%v", request.TeamName, request.UserIDs)

	// валидация
	if request.TeamName == "" {
		log.Printf("Missing team_name")
		writeError(w, "INVALID_REQUEST", "team_name is required", http.StatusBadRequest)
		return
	}

	if len(request.UserIDs) == 0 {
		log.Printf("No users provided")
		writeError(w, "INVALID_REQUEST", "user_ids is required", http.StatusBadRequest)
		return
	}

	// выполняем массовую деактивацию через сервис
	log.Printf("Calling user service for bulk deactivation")
	response, err := h.userService.BulkDeactivateUsers(request.TeamName, request.UserIDs)
	if err != nil {
		log.Printf("Service error: %v", err)
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Code {
			case "NOT_FOUND":
				writeError(w, "NOT_FOUND", serviceErr.Message, http.StatusNotFound)
			case "INVALID_REQUEST":
				writeError(w, "INVALID_REQUEST", serviceErr.Message, http.StatusBadRequest)
			default:
				writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Bulk deactivation completed: %d users deactivated, %d PRs reassigned",
		response.TotalProcessed, response.ReassignedCount)
	writeJSON(w, http.StatusOK, response)
}
