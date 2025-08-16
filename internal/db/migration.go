package db

import (
	"fmt"
	"log"

	"personal-finance/internal/config"

	"database/sql"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4"
)

func MigrateDB(db *sql.DB, cfg *config.Config) error {
	migrationPath := "file://migrations"

	m, err := migrate.New(migrationPath, fmt.Sprintf("sqlite3://%s", cfg.Database.Path))
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migration changes to apply")
			return nil
		}
		return fmt.Errorf("failed to migrate: %w", err)
	}

	log.Println("✅ Database migrated successfully")
	return nil
}