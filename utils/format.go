package utils

import (
	"personal-finance/i18n"
	"personal-finance/internal/config"
	"strings"
)

func FormatAmount(expensesLine string, lang string, cfg *config.Config) string {
	return strings.ReplaceAll(expensesLine, "{{currency}}", i18n.Currency(lang, cfg))
}
