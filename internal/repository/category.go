package repository

import (
	"context"
	"database/sql"
)

func AddCategory(ctx context.Context, db *sql.DB, name string, chatID int64, limit float64) bool {
	_, err := db.ExecContext(ctx, "INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)

	return err == nil
}

func ListCategories(ctx context.Context, db *sql.DB, chatID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM categories WHERE user_id = ?", chatID)
	var categories []string
	if err != nil {
		return categories, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		categories = append(categories, name)
	}

	return categories, nil
}

func CreateCategory(ctx context.Context, db *sql.DB, chatID int64, name string, limit float64) error {
	_, err := db.ExecContext(ctx, "INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)

	return err
}

func NewCategoryName(ctx context.Context, db *sql.DB, chatID int64, name string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories WHERE user_id = ? AND LOWER(name) = ?", chatID, name).Scan(&count)

	return count, err
}

func DeleteCategory(ctx context.Context, db *sql.DB, chatID int64, name string) (int, float64, error) {
	var categoryID int
	var limit float64
	err := db.QueryRowContext(ctx, "SELECT id, limit_sum FROM categories WHERE user_id = ? AND LOWER(name) = ?",
		chatID, name).Scan(&categoryID, &limit)

	return categoryID, limit, err
}

func GetCategory(ctx context.Context, db *sql.DB, chatID int64, categoryInput string) (int, string, error) {
	var categoryID int
	var categoryName string
	err := db.QueryRowContext(ctx, `
        SELECT id, name 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, categoryInput).Scan(&categoryID, &categoryName)

	return categoryID, categoryName, err
}

func GetLimitSum(ctx context.Context, db *sql.DB, chatID int64, categoryID int) (float64, error) {
	var limitSum float64
	err := db.QueryRowContext(ctx, "SELECT limit_sum FROM categories WHERE id = ? AND user_id = ?",
		categoryID, chatID).Scan(&limitSum)

	return limitSum, err
}
