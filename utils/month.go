package utils

import (
	"personal-finance/internal/config"
	"time"
)

func GetMonthName(month time.Time, lang string, cfg *config.Config) string {
	months := map[string][]string{
		"ru": {"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
			"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"},
		"en": {"January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"},
	}

	lang = toValidLang(lang, cfg)
	monthIndex := month.Month() - 1

	return months[lang][monthIndex]
}

func toValidLang(lang string, cfg *config.Config) string {
	for _, supportedLang := range cfg.App.Languages {
		if lang == supportedLang {
			return lang
		}
	}
	return cfg.App.DefaultLanguage
}
