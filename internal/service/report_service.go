package service

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/internal/config"
	"personal-finance/internal/i18n"
	"personal-finance/internal/repository"
	"personal-finance/state"
	"personal-finance/utils"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// generates a report for a single user
func GenerateReport(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, cfg *config.Config) (string, error) {
	now := time.Now()
	month := now.AddDate(0, -1, 0)

	lang, err := state.GetUserLanguage(ctx, redis, chatID)
	if err != nil {
		lang = "en"
	}

	reportRepo := repository.NewReportRepository(db)
	data, err := reportRepo.GetMonthlyReportData(ctx, chatID, month)
	if err != nil {
		return "", err
	}

	if len(data.Categories) == 0 && data.TotalIncome == 0 {
		return "", nil
	}

	monthName := utils.GetMonthName(month, lang, cfg)

	emoji := cfg.App.StatusEmojis.BalanceGood
	if data.Balance < 0 {
		emoji = cfg.App.StatusEmojis.BalanceBad
	} else if data.Balance < data.TotalIncome*(cfg.App.BalanceWarningThreshold/100) {
		emoji = cfg.App.StatusEmojis.BalanceWarning
	}

	var lines []string
	for _, c := range data.Categories {
		statusKey := "monthly_status_under"
		if c.Spent > c.Limit && c.Limit > 0 {
			statusKey = "monthly_status_over"
		}
		status := i18n.T(statusKey, lang, cfg)

		displayName := utils.Title.String(c.Name)
		lines = append(lines, fmt.Sprintf("- %s: %.2f %s / %.2f %s %s",
			displayName,
			c.Spent, cfg.App.CurrencySymbol,
			c.Limit, cfg.App.CurrencySymbol,
			status))
	}

	report := fmt.Sprintf(`%s %s

%s
%s
%s

%s:
%s

%s

%s`,
		emoji, i18n.Tf("monthly_report_title", lang, cfg, monthName),
		i18n.Tf("monthly_income", lang, cfg, utils.FormatAmount(fmt.Sprintf("%.2f", data.TotalIncome), lang, cfg)),
		i18n.Tf("monthly_expenses", lang, cfg, utils.FormatAmount(fmt.Sprintf("%.2f", data.TotalSpent), lang, cfg)),
		i18n.Tf("monthly_balance", lang, cfg, utils.FormatAmount(fmt.Sprintf("%.2f", data.Balance), lang, cfg)),
		i18n.T("monthly_categories", lang, cfg),
		strings.Join(lines, "\n"),
		i18n.Tf("monthly_over_limits", lang, cfg, data.OverLimit),
		i18n.T("monthly_thanks", lang, cfg),
	)

	return report, nil
}
