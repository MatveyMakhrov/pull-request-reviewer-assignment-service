package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// конфигурация подключения к базе данных
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// устанавливает подключение к базе данных с повторными попытками
// принимает: конфигурацию подключения к базе данных
// возвращает: подключение к БД или ошибку после исчерпания попыток
func Connect(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var db *sql.DB
	var err error

	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", attempt, maxAttempts)

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			log.Printf("Database ping failed (attempt %d): %v", attempt, err)
			db.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("Successfully connected to database")

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxAttempts, err)
}
