package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"telegram-shopping-bot/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatalf("TELEGRAM_BOT_TOKEN not set")
	}

	telegramBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	telegramBot.Debug = strings.EqualFold(os.Getenv("TELEGRAM_BOT_DEBUG"), "true")
	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	shoppingBot := bot.NewShoppingBot()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := telegramBot.GetUpdatesChan(u)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log.Printf("Bot started and receiving updates")

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutdown signal received, stopping updates")
			telegramBot.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				log.Printf("Updates channel closed")
				return
			}

			if update.CallbackQuery != nil {
				cb := update.CallbackQuery
				response := bot.HandleCallback(shoppingBot, cb.Data, cb.Message.Chat.ID)

				ack := tgbotapi.NewCallback(cb.ID, response)
				if _, err := telegramBot.Request(ack); err != nil {
					log.Printf("Failed to answer callback: %v", err)
				}

				listMsg := tgbotapi.NewMessage(cb.Message.Chat.ID, shoppingBot.GetList(cb.Message.Chat.ID))
				listMsg.ParseMode = "HTML"
				if keyboard := shoppingBot.BuildListKeyboard(cb.Message.Chat.ID); keyboard != nil {
					listMsg.ReplyMarkup = keyboard
				}
				if _, err := telegramBot.Send(listMsg); err != nil {
					log.Printf("Failed to send list after callback: %v", err)
				}
				continue
			}

			if update.Message == nil {
				continue
			}

			response := bot.HandleUpdate(shoppingBot, update.Message)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			msg.ParseMode = "HTML"
			if update.Message.IsCommand() && update.Message.Command() == "showlist" {
				if keyboard := shoppingBot.BuildListKeyboard(update.Message.Chat.ID); keyboard != nil {
					msg.ReplyMarkup = keyboard
				}
			}
			if _, err := telegramBot.Send(msg); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}
	}
}
