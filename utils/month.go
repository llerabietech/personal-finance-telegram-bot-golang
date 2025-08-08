package utils

import "time"

func GetMonthName(month time.Time, lang string) string {
	months := map[string][]string{
		"ru": {"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
			"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"},
		"en": {"January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"},
	}

	lang = toValidLang(lang)
	monthIndex := month.Month() - 1

	return months[lang][monthIndex]
}

func toValidLang(lang string) string {
	if lang == "ru" {
		return "ru"
	}
	return "en" // default
}
