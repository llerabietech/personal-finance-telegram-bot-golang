package repository

import (
	"context"
	"database/sql"
	"personal-finance/internal/config"
	"time"
)

func AddIncome(ctx context.Context, db *sql.DB, chatID int64, source string, amount float64, cfg *config.Config) error {
	_, err := db.ExecContext(ctx, "INSERT INTO incomes (user_id, source, amount, date) VALUES (?, ?, ?, ?)",
		chatID, source, amount, time.Now().Format(cfg.App.DateFormat))

	return err
}