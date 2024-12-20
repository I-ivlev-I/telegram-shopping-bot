package main

import (
	"log"
	"os"
	"telegram-shopping-bot/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Получаем токен Telegram из окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatalf("TELEGRAM_BOT_TOKEN not set")
	}

	telegramBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	telegramBot.Debug = true
	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	shoppingBot := bot.NewShoppingBot()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramBot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		response := bot.HandleUpdate(shoppingBot, update.Message)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		msg.ParseMode = "HTML"
		if _, err := telegramBot.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}
