package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"pull-request-reviewer-assignment-service/internal/service"
)

// обработчик HTTP запросов для работы с Pull Request'ами
type PRHandler struct {
	prService *service.PRService
}

// создает новый экземпляр обработчика Pull Request'ов с внедрением зависимостей
// принимает: сервис для логики работы с Pull Request'ами
// возвращает: инициализированный обработчик с установленными зависимостями
func NewPRHandler(prService *service.PRService) *PRHandler {
	return &PRHandler{
		prService: prService,
	}
}

// обрабатывает HTTP запрос на создание нового Pull Request с автоназначением ревьюверов
// принимает: HTTP запрос с данными Pull Request и response writer для формирования ответа
// возвращает: JSON ответ с созданным PR или ошибку в случае неудачи
func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received POST /pullRequest/create request")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Invalid JSON: %v", err)
		writeError(w, "INVALID_REQUEST", "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed request: pr_id=%s, name=%s, author=%s",
		request.PullRequestID, request.PullRequestName, request.AuthorID)

	// валидация
	if request.PullRequestID == "" {
		log.Printf("Missing pull_request_id")
		writeError(w, "INVALID_REQUEST", "pull_request_id is required", http.StatusBadRequest)
		return
	}
	if request.PullRequestName == "" {
		log.Printf("Missing pull_request_name")
		writeError(w, "INVALID_REQUEST", "pull_request_name is required", http.StatusBadRequest)
		return
	}
	if request.AuthorID == "" {
		log.Printf("Missing author_id")
		writeError(w, "INVALID_REQUEST", "author_id is required", http.StatusBadRequest)
		return
	}

	// создаем PR через сервис
	log.Printf("Calling PR service to create PR: %s", request.PullRequestID)
	pr, err := h.prService.CreatePR(request.PullRequestID, request.PullRequestName, request.AuthorID)
	if err != nil {
		log.Printf("Service error: %v", err)
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Code {
			case "PR_EXISTS":
				writeError(w, "PR_EXISTS", serviceErr.Message, http.StatusConflict)
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

	log.Printf("PR created successfully: %s", request.PullRequestID)
	response := map[string]interface{}{
		"pr": pr,
	}
	writeJSON(w, http.StatusCreated, response)
}

// обрабатывает запрос на слияние Pull Request
// принимает: HTTP запрос с JSON содержащим pull_request_id
// возвращает: JSON ответ с результатом операции или ошибку
func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received POST /pullRequest/merge request")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Invalid JSON: %v", err)
		writeError(w, "INVALID_REQUEST", "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed request: pr_id=%s", request.PullRequestID)

	// валидация
	if request.PullRequestID == "" {
		log.Printf("Missing pull_request_id")
		writeError(w, "INVALID_REQUEST", "pull_request_id is required", http.StatusBadRequest)
		return
	}

	// мержим PR через сервис
	log.Printf("Calling PR service to merge PR: %s", request.PullRequestID)
	pr, err := h.prService.MergePR(request.PullRequestID)
	if err != nil {
		log.Printf("Service error: %v", err)
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Code {
			case "NOT_FOUND":
				writeError(w, "NOT_FOUND", serviceErr.Message, http.StatusNotFound)
			case "PR_MERGED":
				writeError(w, "PR_MERGED", serviceErr.Message, http.StatusConflict)
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

	log.Printf("PR merged successfully: %s", request.PullRequestID)
	response := map[string]interface{}{
		"pr": pr,
	}
	writeJSON(w, http.StatusOK, response)
}

// переназначает ревьювера в Pull Request на другого пользователя
// принимает: HTTP запрос с JSON содержащим pull_request_id и old_user_id
// возвращает: JSON ответ с обновленным PR и ID нового ревьювера или ошибку
func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received POST /pullRequest/reassign request")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Invalid JSON: %v", err)
		writeError(w, "INVALID_REQUEST", "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed request: pr_id=%s, old_user_id=%s", request.PullRequestID, request.OldUserID)

	// валидация
	if request.PullRequestID == "" {
		log.Printf("Missing pull_request_id")
		writeError(w, "INVALID_REQUEST", "pull_request_id is required", http.StatusBadRequest)
		return
	}
	if request.OldUserID == "" {
		log.Printf("Missing old_user_id")
		writeError(w, "INVALID_REQUEST", "old_user_id is required", http.StatusBadRequest)
		return
	}

	// переназначаем ревьювера через сервис
	log.Printf("Calling PR service to reassign reviewer: %s -> ? in PR: %s", request.OldUserID, request.PullRequestID)
	pr, newReviewerID, err := h.prService.ReassignReviewer(request.PullRequestID, request.OldUserID)
	if err != nil {
		log.Printf("Service error: %v", err)
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Code {
			case "NOT_FOUND":
				writeError(w, "NOT_FOUND", serviceErr.Message, http.StatusNotFound)
			case "PR_MERGED":
				writeError(w, "PR_MERGED", serviceErr.Message, http.StatusConflict)
			case "NOT_ASSIGNED":
				writeError(w, "NOT_ASSIGNED", serviceErr.Message, http.StatusConflict)
			case "NO_CANDIDATE":
				writeError(w, "NO_CANDIDATE", serviceErr.Message, http.StatusConflict)
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

	log.Printf("Reviewer reassigned successfully: %s -> %s in PR: %s", request.OldUserID, newReviewerID, request.PullRequestID)
	response := map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	}
	writeJSON(w, http.StatusOK, response)
}
