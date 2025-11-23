package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"pull-request-reviewer-assignment-service/internal/models"
	"pull-request-reviewer-assignment-service/internal/service"
)

// структура обрабатывает HTTP запросы связанные с управлением командами
type TeamHandler struct {
	teamService *service.TeamService
}

// создает и возвращает новый экземпляр TeamHandler
// принимает: сервис команд для внедрения зависимости
// возвращает: указатель на созданный TeamHandler
func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// создает новую команду с указанными участниками
// принимает: HTTP запрос с JSON содержащим данные команды (название и список участников)
// возвращает: JSON с созданной командой или ошибку валидации/создания
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received POST /team/add request")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		log.Printf("Invalid JSON: %v", err)
		writeError(w, "INVALID_REQUEST", "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed team: %s with %d members", team.TeamName, len(team.Members))

	// валидация
	if team.TeamName == "" {
		log.Printf("Missing team_name")
		writeError(w, "INVALID_REQUEST", "team_name is required", http.StatusBadRequest)
		return
	}

	if len(team.Members) == 0 {
		log.Printf("No members provided")
		writeError(w, "INVALID_REQUEST", "team must have at least one member", http.StatusBadRequest)
		return
	}

	// создаем команду через сервис
	log.Printf("Calling team service to create team: %s", team.TeamName)
	if err := h.teamService.CreateTeam(&team); err != nil {
		log.Printf("Service error: %v", err)
		if serviceErr, ok := err.(*service.ServiceError); ok {
			log.Printf("Service error code: %s, message: %s", serviceErr.Code, serviceErr.Message)
			switch serviceErr.Code {
			case "TEAM_EXISTS":
				writeError(w, "TEAM_EXISTS", serviceErr.Message, http.StatusBadRequest)
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

	log.Printf("Team created successfully: %s", team.TeamName)
	response := map[string]interface{}{
		"team": team,
	}
	writeJSON(w, http.StatusOK, response)
}

// возвращает информацию о команде по её названию
// принимает: HTTP GET запрос с параметром team_name в URL
// возвращает: JSON с данными команды или ошибку если команда не найдена
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received GET /team/get request")

	if r.Method != http.MethodGet {
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, "INVALID_REQUEST", "team_name parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("Getting team: %s", teamName)
	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		if serviceErr, ok := err.(*service.ServiceError); ok && serviceErr.Code == "NOT_FOUND" {
			writeError(w, "NOT_FOUND", serviceErr.Message, http.StatusNotFound)
			return
		}
		writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Team found: %s", teamName)
	response := map[string]interface{}{
		"team": team,
	}
	writeJSON(w, http.StatusOK, response)
}

// вспомогательная функция для отправки JSON ответов
// принимает: ResponseWriter для записи ответа, статус код и данные для сериализации
// возвращает: ничего, просто записывает ответ непосредственно в ResponseWriter
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// вспомогательная функция для отправки ошибок
// принимает: ResponseWriter, код ошибки, сообщение и HTTP статус код
// возвращает: ничего, просто записывает ошибку в ResponseWriter через writeJSON
func writeError(w http.ResponseWriter, errorCode, message string, status int) {
	log.Printf("Error response: %s - %s (status: %d)", errorCode, message, status)
	errorResponse := models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	}
	writeJSON(w, status, errorResponse)
}
