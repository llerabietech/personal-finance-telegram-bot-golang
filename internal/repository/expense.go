package repository

import (
	"context"
	"personal-finance/internal/config"
	"database/sql"
	"time"
)

func AddExpense(ctx context.Context, db *sql.DB, chatID int64, categoryID int, amount float64, cfg *config.Config) error {
	_, err := db.ExecContext(ctx, "INSERT INTO expenses (user_id, category_id, amount, date) VALUES (?, ?, ?, ?)",
		chatID, categoryID, amount, time.Now().Format(cfg.App.DateFormat))

	return err
}

func IsPotentialExpense(ctx context.Context, db *sql.DB, chatID int64, categoryName string) bool {
	var count int
	db.QueryRowContext(ctx, `
        SELECT COUNT(*) 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, categoryName).Scan(&count)

	return count > 0
}

func GetSpent(ctx context.Context, db *sql.DB, chatID int64, categoryID int, month string) (float64, error) {
	var spent float64
	err := db.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(amount), 0)
        FROM expenses
        WHERE user_id = ? AND category_id = ? AND date LIKE ?`,
		chatID, categoryID, month+"%").Scan(&spent)

	return spent, err
}