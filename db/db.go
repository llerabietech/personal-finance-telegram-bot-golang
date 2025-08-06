package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./finance.db")
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
    );`

	_, err := DB.Exec(query)
	if err != nil {
		panic(err)
	}
}

func GetAllUsers() ([]int64, error) {
    rows, err := DB.Query("SELECT DISTINCT user_id FROM expenses WHERE date >= date('now', '-3 month')")
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
