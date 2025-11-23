package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pull-request-reviewer-assignment-service/internal/config"
	"pull-request-reviewer-assignment-service/internal/database"
	"pull-request-reviewer-assignment-service/internal/handlers"
	"pull-request-reviewer-assignment-service/internal/repository"
	"pull-request-reviewer-assignment-service/internal/repository/postgres"
	"pull-request-reviewer-assignment-service/internal/service"
	"syscall"
	"time"
)

func main() {
	// загрузка конфигурации
	cfg := config.Load()

	log.Println("PR Reviewer Service Starting...")
	log.Printf("Port: %s", cfg.ServerPort)
	log.Printf("Database: %s@%s:%s/%s",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// подключаемся к базе данных
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Database not available - cannot start without database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// применяем миграции
	if err := database.SimpleRunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations applied successfully")

	// инициализируем репозитории
	var teamRepo repository.TeamRepository
	var userRepo repository.UserRepository
	var prRepo repository.PRRepository
	var reviewRepo repository.ReviewRepository
	var statsRepo repository.StatsRepository

	if db != nil {
		// используем PostgreSQL репозитории
		teamRepo = postgres.NewTeamRepository(db)
		userRepo = postgres.NewUserRepository(db)
		prRepo = postgres.NewPRRepository(db)
		reviewRepo = postgres.NewReviewRepository(db)
		statsRepo = postgres.NewStatsRepository(db)
		log.Println("Using PostgreSQL repositories")
	} else {
		log.Println("Database not available - cannot start without database")
		return
	}

	// инициализируем сервисы
	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo, prRepo, teamRepo, reviewRepo)
	prService := service.NewPRService(prRepo, reviewRepo, userRepo, teamService)
	statsService := service.NewStatsService(statsRepo)

	// инициализируем ручки
	teamHandler := handlers.NewTeamHandler(teamService)
	userHandler := handlers.NewUserHandler(userService)
	prHandler := handlers.NewPRHandler(prService)
	statsHandler := handlers.NewStatsHandler(statsService)

	mux := http.NewServeMux()

	// регистрируем ручки
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/team/add", teamHandler.AddTeam)
	mux.HandleFunc("/team/get", teamHandler.GetTeam)
	mux.HandleFunc("/users/setIsActive", userHandler.SetUserActive)
	mux.HandleFunc("/pullRequest/create", prHandler.CreatePR)
	mux.HandleFunc("/pullRequest/merge", prHandler.MergePR)
	mux.HandleFunc("/pullRequest/reassign", prHandler.ReassignReviewer)
	mux.HandleFunc("/users/getReview", userHandler.GetUserReviewPRs)
	mux.HandleFunc("/stats/review-assignments", statsHandler.GetReviewStats)
	mux.HandleFunc("/users/bulk-deactivate", userHandler.BulkDeactivate)
	mux.HandleFunc("/", homeHandler)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	// логируем эндпоинты
	go func() {
		log.Println("Server is ready to handle requests")
		log.Println("Available endpoints:")
		log.Println("   GET  /health")
		log.Println("   POST /team/add")
		log.Println("   GET  /team/get?team_name=...")
		log.Println("   POST /users/setIsActive")
		log.Println("   POST /pullRequest/create")
		log.Println("   POST /pullRequest/merge")
		log.Println("   POST /pullRequest/reassign")
		log.Println("   GET  /users/getReview?user_id=...")
		log.Println("   GET  /stats/review-assignments")
		log.Println("   POST /users/bulk-deactivate")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// обработчик эндпоинта проверки healthy сервиса
// принимает: HTTP запрос и writer для ответа на запросы проверки health
// возвращает: JSON ответ со статусом, названием и версией сервиса
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse := `{"status":"healthy","service":"PR Reviewer Assignment Service","version":"1.0.0"}`
	w.Write([]byte(jsonResponse))
}

// обработчик корневого эндпоинт
// принимает: HTTP запрос и writer для ответа на запросы к корневому пути
// возвращает: JSON с описанием сервиса, версией и списком доступных эндпоинтов
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := `{
		"service": "PR Reviewer Service is running!",
		"version": "1.0.0",
		"endpoints": {
			"health": "/health",
			"teams": "/team/add, /team/get",
			"users": "/users/setIsActive, /users/getReview",
			"pull_requests": "/pullRequest/create, /pullRequest/merge, /pullRequest/reassign"
		}
	}`

	w.Write([]byte(response))
}
