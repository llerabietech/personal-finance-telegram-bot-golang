package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"os"
	"personal-finance/bot"
	"personal-finance/db"
)

type Config struct {
	TelegramBotToken string
	RedisPassword    string
}

var TitleCaser = cases.Title(language.Und)

func main() {
	telegramBotToken := os.Getenv("TELEGRAM_TOKEN")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	db.InitDB()
	db.InitRedis(redisPassword)

	botApi, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	botApi.Debug = false
	log.Printf("Authorized on account %s", botApi.Self.UserName)

	bot.StartScheduler(botApi)
	bot.StartBot(botApi)
}
