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
	removeLegacyKeyboard := func(chatID int64) {
		// Telegram only removes a persistent reply keyboard through a sent
		// message. Delete that service message immediately so migration from the
		// old keyboard does not leave an extra reply in the chat.
		cleanup := tgbotapi.NewMessage(chatID, "⌨️ Обновление меню…")
		cleanup.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
		sent, err := telegramBot.Send(cleanup)
		if err != nil {
			log.Printf("Failed to remove legacy reply keyboard: %v", err)
			return
		}
		if _, err := telegramBot.Request(tgbotapi.NewDeleteMessage(chatID, sent.MessageID)); err != nil {
			log.Printf("Failed to delete keyboard migration message: %v", err)
		}
	}

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
				if cb.Message == nil {
					ack := tgbotapi.NewCallbackWithAlert(cb.ID, "Не удалось определить сообщение меню.")
					if _, err := telegramBot.Request(ack); err != nil {
						log.Printf("Failed to answer callback: %v", err)
					}
					continue
				}

				response := bot.HandleCallback(shoppingBot, cb.Data, cb.Message.Chat.ID)
				if strings.HasPrefix(cb.Data, "del:") || strings.HasPrefix(cb.Data, "str:") || strings.HasPrefix(cb.Data, "uns:") {
					response += "\n\n" + shoppingBot.GetList(cb.Message.Chat.ID)
				}

				ack := tgbotapi.NewCallback(cb.ID, "")
				if _, err := telegramBot.Request(ack); err != nil {
					log.Printf("Failed to answer callback: %v", err)
				}

				markup := bot.MainMenuKeyboard()
				if bot.CallbackShowsList(cb.Data) {
					markup = *shoppingBot.BuildListKeyboard(cb.Message.Chat.ID)
				}
				// Edit the menu message in place. Sending a new message for every
				// button press made the chat jump to an apparent reply and filled it
				// with duplicate menus.
				edit := tgbotapi.NewEditMessageTextAndMarkup(cb.Message.Chat.ID, cb.Message.MessageID, response, markup)
				edit.ParseMode = "HTML"
				if _, err := telegramBot.Send(edit); err != nil && !strings.Contains(err.Error(), "message is not modified") {
					log.Printf("Failed to update callback message: %v", err)
				}
				continue
			}

			if update.Message == nil {
				continue
			}
			if update.Message.IsCommand() && update.Message.Command() == "start" {
				removeLegacyKeyboard(update.Message.Chat.ID)
			}

			response := bot.HandleUpdate(shoppingBot, update.Message)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			msg.ParseMode = "HTML"
			menu := bot.MainMenuKeyboard()
			msg.ReplyMarkup = menu
			if (update.Message.IsCommand() && update.Message.Command() == "showlist") || update.Message.Text == bot.BtnShowList {
				msg.ReplyMarkup = shoppingBot.BuildListKeyboard(update.Message.Chat.ID)
			}
			if _, err := telegramBot.Send(msg); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}
	}
}
