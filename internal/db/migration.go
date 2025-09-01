package db

import (
	"fmt"
	"log"
	"strings"

	"personal-finance/internal/config"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateDB(db *sql.DB, cfg *config.Config) error {
	sourceURL := cfg.Database.MigrationsPath
	if !strings.HasPrefix(sourceURL, "file://") {
		sourceURL = "file://" + sourceURL
	}

	m, err := migrate.New(sourceURL, fmt.Sprintf("sqlite3://%s", cfg.Database.Path))
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
