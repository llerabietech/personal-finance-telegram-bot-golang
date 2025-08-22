package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	"personal-finance/internal/config"
	"personal-finance/internal/i18n"
	"personal-finance/internal/repository"
	"personal-finance/state"
	"personal-finance/utils"
	"strconv"
	"strings"
)

func AddCategory(ctx context.Context, db *sql.DB, chatID int64, input string, lang string, cfg *config.Config) string {
	parts := strings.Fields(input)

	name := strings.ToLower(parts[1])
	limit, err := strconv.ParseFloat(parts[2], 64)
	if err != nil || limit <= 0 {
		return i18n.T("correct_digit", lang, cfg)
	}

	result := repository.AddCategory(ctx, db, name, chatID, limit)
	if !result {
		return i18n.T("error_add_category", lang, cfg)
	}

	return utils.FormatAmount(i18n.Tf("category_created", lang, cfg, name, limit), lang, cfg)
}

func ListCategories(ctx context.Context, db *sql.DB, chatID int64, lang string, cfg *config.Config) string {
	categories, err := repository.ListCategories(ctx, db, chatID)
	if err != nil {
		return i18n.T("error_load_category", lang, cfg)
	}

	if len(categories) == 0 {
		return i18n.T("error_empty_category", lang, cfg)
	}
	return i18n.T("categories2", lang, cfg) + strings.Join(categories, "\n• ")
}

func CreateCategory(ctx context.Context, db *sql.DB, chatID int64, name string, limit float64, lang string, cfg *config.Config) string {
	err := repository.CreateCategory(ctx, db, chatID, name, limit)
	if err != nil {
		return i18n.T("error_create_category", lang, cfg)
	}
	return utils.FormatAmount(i18n.Tf("category_created", lang, cfg, utils.Title.String(name), limit), lang, cfg)
}

func NewCategoryName(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, text string, lang string, cfg *config.Config) string {
	name := strings.ToLower(strings.TrimSpace(text))
	if name == "" {
		return i18n.T("empty_name", lang, cfg)
	}

	count, err := repository.NewCategoryName(ctx, db, chatID, name)
	if err != nil {
		return err.Error()
	}
	if count > 0 {
		state.Clear(ctx, redis, chatID)
		return i18n.T("error_category_already_exist", lang, cfg)
	}

	state.SetTempData(ctx, redis, chatID, name)
	state.SetState(ctx, redis, chatID, state.AwaitingCategoryLimit)

	return i18n.Tf("enter_limit2", lang, cfg, utils.Title.String(name))
}

func DeleteCategory(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, categoryName string, lang string, cfg *config.Config) string {
	name := strings.ToLower(strings.TrimSpace(categoryName))

	categoryID, limit, err := repository.DeleteCategory(ctx, db, chatID, name)

	if err == sql.ErrNoRows {
		return i18n.T("category_not_found2", lang, cfg)
	} else if err != nil {
		return i18n.T("error_found_category", lang, cfg)
	}

	state.SetTempData(ctx, redis, chatID, fmt.Sprintf("%d", categoryID))
	state.SetState(ctx, redis, chatID, state.ConfirmDeleteCategory)

	displayName := utils.Title.String(name)
	return utils.FormatAmount(i18n.Tf("confirm_delete", lang, cfg, displayName, limit), lang, cfg)
}
