package db

import (
	"fmt"
	"strings"

	"personal-finance/internal/config"
	"personal-finance/internal/log"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

func MigrateDB(db *sql.DB, cfg *config.Config) error {
	log.Info("Starting database migration")

	sourceURL := cfg.Database.MigrationsPath
	if !strings.HasPrefix(sourceURL, "file://") {
		sourceURL = "file://" + sourceURL
	}

	log.WithFields(logrus.Fields{
		"source_url": sourceURL,
		"db_path":    cfg.Database.Path,
	}).Debug("Migration configuration")

	m, err := migrate.New(sourceURL, fmt.Sprintf("sqlite3://%s", cfg.Database.Path))
	if err != nil {
		log.WithError(err).Error("Failed to create migrate instance")
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("No migration changes to apply")
			return nil
		}
		log.WithError(err).Error("Migration failed")
		return fmt.Errorf("failed to migrate: %w", err)
	}

	log.Info("Database migrated successfully")
	return nil
}
