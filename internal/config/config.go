package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var globalConfig *Config

type Config struct {
	Telegram TelegramConfig
	Database DatabaseConfig
	Redis    RedisConfig
	App      AppConfig
	Logging  LoggingConfig
}

type TelegramConfig struct {
	BotToken string
	Debug    bool
}

type DatabaseConfig struct {
	Path           string
	MigrationsPath string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AppConfig struct {
	Timezone                string
	CleanupMonths           int
	ReportHour              int
	ReportMinute            int
	LimitWarningThreshold   float64      // Warning threshold for spending limits (default: 80%)
	LimitOverloadThreshold  float64      // Overload threshold for spending limits (default: 100%)
	BalanceWarningThreshold float64      // Balance warning threshold (default: 10%)
	DateFormat              string       // Date format (default: "2006-01-02")
	MonthFormat             string       // Month format (default: "2006-01")
	TimeFormat              string       // Time format (default: "15:04")
	CurrencySymbol          string       // Currency symbol (default: "₽")
	StatusEmojis            StatusEmojis // Status emojis
	ConfirmationWords       []string     // Confirmation words (default: ["да", "yes"])
	Languages               []string     // Supported languages (default: ["ru", "en"])
	DefaultLanguage         string       // Default language (default: "en")
}

type StatusEmojis struct {
	Success        string // Success emoji (default: "✅")
	Warning        string // Warning emoji (default: "🟡")
	Error          string // Error emoji (default: "❌")
	BalanceGood    string // Good balance emoji (default: "🟢")
	BalanceWarning string // Balance warning emoji (default: "🟡")
	BalanceBad     string // Bad balance emoji (default: "🔴")
}

type LoggingConfig struct {
	Level  string // Log level: debug, info, warn, error, fatal, panic
	Format string // Log format: text, json
}

func Load() (*Config, error) {
	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken: getEnvOrDefault("TELEGRAM_TOKEN", ""),
			Debug:    getEnvBoolOrDefault("TELEGRAM_DEBUG", false),
		},
		Database: DatabaseConfig{
			Path:           getEnvOrDefault("DB_PATH", "./finance.db"),
			MigrationsPath: getEnvOrDefault("MIGRATIONS_PATH", "migrations"),
		},
		Redis: RedisConfig{
			Addr:     getEnvOrDefault("REDIS_ADDR", "redis:6379"),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       getEnvIntOrDefault("REDIS_DB", 0),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "text"),
		},
		App: AppConfig{
			Timezone:                getEnvOrDefault("TZ", "Europe/Moscow"),
			CleanupMonths:           getEnvIntOrDefault("CLEANUP_MONTHS", 3),
			ReportHour:              getEnvIntOrDefault("REPORT_HOUR", 0),
			ReportMinute:            getEnvIntOrDefault("REPORT_MINUTE", 0),
			LimitWarningThreshold:   getEnvFloatOrDefault("LIMIT_WARNING_THRESHOLD", 80.0),
			LimitOverloadThreshold:  getEnvFloatOrDefault("LIMIT_OVERLOAD_THRESHOLD", 100.0),
			BalanceWarningThreshold: getEnvFloatOrDefault("BALANCE_WARNING_THRESHOLD", 10.0),
			DateFormat:              getEnvOrDefault("DATE_FORMAT", "2006-01-02"),
			MonthFormat:             getEnvOrDefault("MONTH_FORMAT", "2006-01"),
			TimeFormat:              getEnvOrDefault("TIME_FORMAT", "15:04"),
			CurrencySymbol:          getEnvOrDefault("CURRENCY_SYMBOL", "₽"),
			StatusEmojis: StatusEmojis{
				Success:        getEnvOrDefault("EMOJI_SUCCESS", "✅"),
				Warning:        getEnvOrDefault("EMOJI_WARNING", "🟡"),
				Error:          getEnvOrDefault("EMOJI_ERROR", "❌"),
				BalanceGood:    getEnvOrDefault("EMOJI_BALANCE_GOOD", "🟢"),
				BalanceWarning: getEnvOrDefault("EMOJI_BALANCE_WARNING", "🟡"),
				BalanceBad:     getEnvOrDefault("EMOJI_BALANCE_BAD", "🔴"),
			},
			ConfirmationWords: getEnvStringSliceOrDefault("CONFIRMATION_WORDS", []string{"да", "yes"}),
			Languages:         getEnvStringSliceOrDefault("LANGUAGES", []string{"ru", "en"}),
			DefaultLanguage:   getEnvOrDefault("DEFAULT_LANGUAGE", "en"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

func Get() *Config {
	if globalConfig == nil {
		return nil
	}
	return globalConfig
}

func (c *Config) validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	if c.Database.Path == "" {
		return fmt.Errorf("DB_PATH cannot be empty")
	}

	if c.Redis.Addr == "" {
		return fmt.Errorf("REDIS_ADDR cannot be empty")
	}

	if c.App.CleanupMonths < 1 {
		return fmt.Errorf("CLEANUP_MONTHS must be at least 1")
	}

	if c.App.ReportHour < 0 || c.App.ReportHour > 23 {
		return fmt.Errorf("REPORT_HOUR must be between 0 and 23")
	}

	if c.App.ReportMinute < 0 || c.App.ReportMinute > 59 {
		return fmt.Errorf("REPORT_MINUTE must be between 0 and 59")
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
