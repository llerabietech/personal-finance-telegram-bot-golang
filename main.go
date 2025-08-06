package main

import (
    "fmt"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "log"
    "os"
    "encoding/json"
)

type Config struct {
    TelegramBotToken string
}

func main() {
    file, _ := os.Open("configs/config.json")
    decoder := json.NewDecoder(file)
    configuration := Config{}
    err := decoder.Decode(&configuration)
    if err != nil {
       log.Panic(err)
    }
    fmt.Println(configuration.TelegramBotToken)

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}