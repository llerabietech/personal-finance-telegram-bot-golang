package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"personal-finance/internal/config"

	"gopkg.in/yaml.v3"
)

var CurrencySymbols = map[string]string{
	"ru": "₽",
	"en": "$",
	//TODO
	// "es": "€",
}

var Translations map[string]map[string]string

func LoadTranslations() error {
	Translations = make(map[string]map[string]string)

	var localesPath string

	if _, err := os.Stat("internal/i18n/locales"); err == nil {
		localesPath = "internal/i18n/locales"
	} else if _, err := os.Stat("locales"); err == nil {
		localesPath = "locales"
	} else {
		return fmt.Errorf("locales directory not found")
	}

	files, err := os.ReadDir(localesPath)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		lang := file.Name()[:len(file.Name())-5] // del ".yaml"

		filePath := filepath.Join(localesPath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		var translations map[string]string
		if err := yaml.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
		}

		Translations[lang] = translations
	}

	return nil
}

func T(key, lang string, cfg *config.Config) string {
	lang = toValidLang(lang, cfg)
	if text, exists := Translations[lang][key]; exists {
		return text
	}
	return key
}

func Tf(key, lang string, cfg *config.Config, args ...interface{}) string {
	return fmt.Sprintf(T(key, lang, cfg), args...)
}

func toValidLang(lang string, cfg *config.Config) string {
	for _, supportedLang := range cfg.App.Languages {
		if lang == supportedLang {
			return lang
		}
	}
	return cfg.App.DefaultLanguage
}

func Currency(lang string, cfg *config.Config) string {
	if symbol, exists := CurrencySymbols[lang]; exists {
		return symbol
	}
	return cfg.App.CurrencySymbol // default
}
