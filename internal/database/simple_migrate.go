package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// применяет миграции базы данных из папки migrations с проверкой существования таблиц
// принимает: подключение к базе данных для применения SQL-миграций
// возвращает: ошибку в случае неудачи или nil при успешном выполнении/пропуске миграций
func SimpleRunMigrations(db *sql.DB) error {
	migrationsPath := "migrations"

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Printf("Migrations directory does not exist: %s", migrationsPath)
		return nil
	}

	log.Println("Checking database state...")
	tablesExist, err := checkIfTablesExist(db)
	if err != nil {
		return fmt.Errorf("failed to check database state: %w", err)
	}

	if tablesExist {
		log.Println("Database tables already exist, skipping migrations")
		return nil
	}

	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("could not read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") && !strings.Contains(file.Name(), "down.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Strings(migrationFiles)

	if len(migrationFiles) == 0 {
		log.Println("No migration files found")
		return nil
	}

	log.Println("Starting database migrations...")

	for _, filename := range migrationFiles {
		filePath := filepath.Join(migrationsPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not read migration file %s: %w", filename, err)
		}

		log.Printf("Applying migration: %s", filename)

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("could not execute migration %s: %w", filename, err)
		}

		log.Printf("Applied migration: %s", filename)
	}

	log.Println("All migrations applied successfully")
	return nil
}

// проверяет существование всех основных таблиц базы данных
// принимает: подключение к базе данных для выполнения проверочных запросов
// возвращает: true если все таблицы существуют или false если отсутствует хотя бы одна таблица
func checkIfTablesExist(db *sql.DB) (bool, error) {
	requiredTables := []string{"teams", "users", "pull_requests", "pr_reviewers"}

	for _, table := range requiredTables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("failed to check if table %s exists: %w", table, err)
		}

		if !exists {
			log.Printf("Table %s does not exist", table)
			return false, nil
		}

		log.Printf("Table %s exists", table)
	}

	return true, nil
}
