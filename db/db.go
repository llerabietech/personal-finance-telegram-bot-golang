package db

import (
	"database/sql"
	"personal-finance/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(cfg *config.Config) {
	var err error
	DB, err = sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		panic(err)
	}

	createTables()
}

func createTables() {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY
    );
    CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        user_id INTEGER,
        limit_sum REAL,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
    CREATE TABLE IF NOT EXISTS expenses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        category_id INTEGER,
        amount REAL,
        date TEXT,
        FOREIGN KEY(category_id) REFERENCES categories(id)
    );
    CREATE TABLE IF NOT EXISTS incomes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        source TEXT,
        amount REAL,
        date TEXT
    );`

	_, err := DB.Exec(query)
	if err != nil {
		panic(err)
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
