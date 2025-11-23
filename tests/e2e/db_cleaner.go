package e2e

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	testDBHost     = "localhost"
	testDBPort     = 5434
	testDBUser     = "postgres"
	testDBPassword = "password"
	testDBName     = "pr_reviewer_e2e"
)

// CleanTestDatabase полностью очищает тестовую БД
func CleanTestDatabase() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}
	defer db.Close()

	// Очищаем таблицы в правильном порядке из-за foreign keys
	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Printf("Warning: failed to truncate table %s: %v", table, err)
		}
	}

	return nil
}
