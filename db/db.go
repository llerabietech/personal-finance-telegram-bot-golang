package db

import (
	"database/sql"
	"log"
	"personal-finance/internal/config"
	"personal-finance/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(cfg *config.Config) {
	var err error
	DB, err = sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		panic(err)
	}

	if err := db.MigrateDB(DB, cfg); err != nil {
		log.Fatal("Migration failed: ", err)
	}
}

func GetActiveUsersLastQuarter(cfg *config.Config) ([]int64, error) {
	rows, err := DB.Query("SELECT DISTINCT user_id FROM expenses WHERE date >= date('now', '-? month')", cfg.App.CleanupMonths)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			continue
		}
		users = append(users, chatID)
	}
	return users, nil
}
