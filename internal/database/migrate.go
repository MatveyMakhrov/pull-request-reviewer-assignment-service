package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// применяет миграции к базе данных
// принимает: подключение к БД в качестве параметра
// возвращает: ошибку в случае неудачи или nil при успешном выполнении
func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	// получаем абсолютный путь к миграциям
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return fmt.Errorf("could not get absolute path to migrations: %w", err)
	}

	migrationSource := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.NewWithDatabaseInstance(
		migrationSource,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %w", err)
	}

	// применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("could not get migration version: %w", err)
	}

	log.Printf("Migrations applied successfully. Version: %d, Dirty: %t", version, dirty)
	return nil
}
