package i18n

import (
	"personal-finance/internal/config"
	"testing"
)

func TestLoadTranslations(t *testing.T) {
	err := LoadTranslations()
	if err != nil {
		t.Fatalf("Failed to load translations: %v", err)
	}

	if len(Translations) == 0 {
		t.Fatal("No translations loaded")
	}

	if _, exists := Translations["ru"]; !exists {
		t.Fatal("Russian translations not found")
	}

	if _, exists := Translations["en"]; !exists {
		t.Fatal("English translations not found")
	}

	ruTranslations := Translations["ru"]
	enTranslations := Translations["en"]

	if ruTranslations["welcome"] != "Привет! Выбери язык:" {
		t.Errorf("Russian welcome translation mismatch: got %s", ruTranslations["welcome"])
	}

	if enTranslations["welcome"] != "Hi! Please choose your language:" {
		t.Errorf("English welcome translation mismatch: got %s", enTranslations["welcome"])
	}

	if len(ruTranslations) < 50 {
		t.Errorf("Expected at least 50 Russian translations, got %d", len(ruTranslations))
	}

	if len(enTranslations) < 50 {
		t.Errorf("Expected at least 50 English translations, got %d", len(enTranslations))
	}
}

func TestTFunction(t *testing.T) {
	err := LoadTranslations()
	if err != nil {
		t.Fatalf("Failed to load translations: %v", err)
	}

	mockConfig := &config.Config{
		App: config.AppConfig{
			Languages:       []string{"ru", "en"},
			DefaultLanguage: "en",
		},
	}

	ruText := T("welcome", "ru", mockConfig)
	if ruText != "Привет! Выбери язык:" {
		t.Errorf("Expected Russian welcome, got: %s", ruText)
	}

	enText := T("welcome", "en", mockConfig)
	if enText != "Hi! Please choose your language:" {
		t.Errorf("Expected English welcome, got: %s", enText)
	}

	unknownText := T("unknown_key", "ru", mockConfig)
	if unknownText != "unknown_key" {
		t.Errorf("Expected unknown key to be returned as-is, got: %s", unknownText)
	}
}

func TestTfFunction(t *testing.T) {
	err := LoadTranslations()
	if err != nil {
		t.Fatalf("Failed to load translations: %v", err)
	}

	mockConfig := &config.Config{
		App: config.AppConfig{
			Languages:       []string{"ru", "en"},
			DefaultLanguage: "en",
		},
	}

	ruText := Tf("category_created", "ru", mockConfig, "Еда", 1000.0)
	expectedRu := "✅ Категория 'Еда' добавлена с лимитом 1000.00 {{currency}}"
	if ruText != expectedRu {
		t.Errorf("Expected Russian formatted text, got: %s", ruText)
	}

	enText := Tf("category_created", "en", mockConfig, "Food", 1000.0)
	expectedEn := "✅ Category 'Food' added with limit 1000.00 {{currency}}"
	if enText != expectedEn {
		t.Errorf("Expected English formatted text, got: %s", enText)
	}
}
