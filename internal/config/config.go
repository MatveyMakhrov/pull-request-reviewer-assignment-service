package config

import (
	"os"
	"pull-request-reviewer-assignment-service/internal/database"
)

// структура приложения, содержащая настройки сервера и базы данных
type Config struct {
	ServerPort string
	Database   database.Config
}

// загружает структуру приложения из переменных окружения с значениями по умолчанию
// принимает: значения из переменных окружения или использует значения по умолчанию
// возвращает: указатель на структуру Config с настройками сервера и базы данных
func Load() *Config {
	return &Config{
		ServerPort: getEnv("PORT", "8080"),
		Database: database.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "pr_reviewer"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

// получает значение переменной окружения или возвращает значение по умолчанию
// принимает: ключ переменной окружения и значение по умолчанию
// возвращает: значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
