package utils

import (
	"strings"
	"personal-finance/i18n"
)

func FormatAmount(expensesLine string, lang string) string {
    return strings.ReplaceAll(expensesLine, "{{currency}}", i18n.Currency(lang))
}