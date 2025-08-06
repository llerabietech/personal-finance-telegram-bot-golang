package main

import (
	"encoding/json"
	"log"
	"os"
	"personal-finance/db"
	"personal-finance/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	TelegramBotToken string
	RedisPassword string
}

func main() {
	file, _ := os.Open("configs/config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}

	db.InitDB()
	db.InitRedis(configuration.RedisPassword)
	
	botApi, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	botApi.Debug = false
	log.Printf("Authorized on account %s", botApi.Self.UserName)

	bot.StartScheduler(botApi)
	bot.StartBot(botApi)
}
