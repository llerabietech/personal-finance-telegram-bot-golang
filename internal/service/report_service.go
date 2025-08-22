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

func GetAnalytics(ctx context.Context, db *sql.DB, chatID int64, lang string, cfg *config.Config) string {
	month := time.Now().Format(cfg.App.MonthFormat)

	rows, totalIncome, totalExpenses, err := repository.GetAnalyticsData(ctx, db, chatID, month)
	if err != nil {
		return i18n.T("limits2", lang, cfg)
	}
	defer rows.Close()

	var report []string
	for rows.Next() {
		var name string
		var spent, limit float64
		rows.Scan(&name, &spent, &limit)
		status := cfg.App.StatusEmojis.Success
		if spent > limit {
			status = cfg.App.StatusEmojis.Error
		}
		report = append(report, fmt.Sprintf("• %s: %.2f %s / %.2f %s %s", name, spent, cfg.App.CurrencySymbol, limit, cfg.App.CurrencySymbol, status))
	}

	balance := totalIncome - totalExpenses
	balanceEmoji := cfg.App.StatusEmojis.BalanceGood
	if balance < 0 {
		balanceEmoji = cfg.App.StatusEmojis.BalanceBad
	} else if balance < totalIncome*(cfg.App.BalanceWarningThreshold/100) {
		balanceEmoji = cfg.App.StatusEmojis.BalanceWarning
	}

	details := "—"
	if len(report) > 0 {
		details = strings.Join(report, "\n")
	}

	return utils.FormatAmount(i18n.Tf("analytics_title", lang, cfg, utils.GetMonthName(time.Now(), lang, cfg), totalIncome, totalExpenses, balanceEmoji, balance, details), lang, cfg)
}
