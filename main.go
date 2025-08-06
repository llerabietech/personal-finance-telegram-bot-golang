package main

import (
	"encoding/json"
	"log"
	"os"
	"personal-finance/bot"
	"personal-finance/db"
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
	
	bot.StartBot(configuration.TelegramBotToken)
}
