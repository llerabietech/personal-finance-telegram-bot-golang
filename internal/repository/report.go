package repository

import (
	"context"
	"database/sql"
	"personal-finance/internal/config"
	"time"
)

type CategoryReport struct {
	Name  string
	Spent float64
	Limit float64
}

type ReportData struct {
	TotalSpent  float64
	TotalIncome float64
	Balance     float64
	Categories  []CategoryReport
	OverLimit   int
}

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetMonthlyReportData(ctx context.Context, chatID int64, month time.Time) (*ReportData, error) {
	monthStr := month.Format("2006-01")

	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			c.name,
			SUM(e.amount),
			COALESCE(c.limit_sum, 0)
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = ? AND e.date LIKE ?
		GROUP BY c.name, c.limit_sum
	`, chatID, monthStr+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []CategoryReport
	var totalSpent, totalIncome float64
	var overLimit int

	for rows.Next() {
		var c CategoryReport
		if err := rows.Scan(&c.Name, &c.Spent, &c.Limit); err != nil {
			return nil, err
		}
		totalSpent += c.Spent
		if c.Spent > c.Limit && c.Limit > 0 {
			overLimit++
		}
		categories = append(categories, c)
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0) 
		FROM incomes 
		WHERE user_id = ? AND date LIKE ?
	`, chatID, monthStr+"%").Scan(&totalIncome)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &ReportData{
		TotalSpent:  totalSpent,
		TotalIncome: totalIncome,
		Balance:     totalIncome - totalSpent,
		Categories:  categories,
		OverLimit:   overLimit,
	}, nil
}

func CleanupOldExpenses(ctx context.Context, db *sql.DB, cfg *config.Config) {
	threeMonthsAgo := time.Now().AddDate(0, -cfg.App.CleanupMonths, 0).Format(cfg.App.DateFormat)

	result, err := db.ExecContext(ctx, "DELETE FROM expenses WHERE date < ?", threeMonthsAgo)
	if err != nil {
		println("Error when deleting old expenses:", err.Error())
		return
	}

	rows, _ := result.RowsAffected()
	println("Cleanup: deleted", rows, "old expenses (before", threeMonthsAgo+")")
}

func GetActiveUsersLastQuarter(ctx context.Context, db *sql.DB, cfg *config.Config) ([]int64, error) {
	rows, err := db.QueryContext(ctx, "SELECT DISTINCT user_id FROM expenses WHERE date >= date('now', '-? month')", cfg.App.CleanupMonths)
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

func GetAnalyticsData(ctx context.Context, db *sql.DB, chatID int64, month string) (*sql.Rows, float64, float64, error) {
	var totalIncome float64
	err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE user_id = ? AND date LIKE ?",
		chatID, month+"%").Scan(&totalIncome)
	if err != nil {
		totalIncome = 0
	}

	var totalExpenses float64
	err = db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM expenses e JOIN categories c ON e.category_id = c.id WHERE e.user_id = ? AND e.date LIKE ?",
		chatID, month+"%").Scan(&totalExpenses)
	if err != nil {
		totalExpenses = 0
	}

	rows, err := db.QueryContext(ctx, `
        SELECT c.name, SUM(e.amount), c.limit_sum 
        FROM expenses e
        JOIN categories c ON e.category_id = c.id
        WHERE e.user_id = ? AND e.date LIKE ?
        GROUP BY c.name, c.limit_sum`, chatID, month+"%")

	return rows, totalIncome, totalExpenses, err
}
