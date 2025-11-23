package handlers

import (
	"log"
	"net/http"
	"pull-request-reviewer-assignment-service/internal/service"
)

// структура обрабатывает HTTP запросы для получения статистики
type StatsHandler struct {
	statsService *service.StatsService
}

// создает и возвращает новый экземпляр StatsHandler
// принимает: сервис статистики для внедрения зависимости
// возвращает: указатель на созданный StatsHandler
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// возвращает статистику по назначениям на код-ревью
// принимает: HTTP GET запрос без параметров
// возвращает: JSON со статистикой назначений или ошибку
func (h *StatsHandler) GetReviewStats(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received GET /stats/review-assignments request")

	if r.Method != http.MethodGet {
		writeError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.statsService.GetReviewStats()
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		writeError(w, "INTERNAL_ERROR", "Failed to retrieve statistics", http.StatusInternalServerError)
		return
	}

	log.Printf("Statistics retrieved: %d total assignments", stats.TotalAssignments)
	writeJSON(w, http.StatusOK, stats)
}
