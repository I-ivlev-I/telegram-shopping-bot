package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

//Заменил на использование .env файла перед выгрузкой в GitHub
//const YOUR_TELEGRAM_BOT_TOKEN = "YOUR_TELEGRAM_BOT_TOKEN"

// Структура для хранения списка покупок на чат
type ShoppingBot struct {
	mu            sync.Mutex
	shoppingLists map[int64][]string // ключ - chat ID, значение - список покупок
}

func NewShoppingBot() *ShoppingBot {
	return &ShoppingBot{
		shoppingLists: make(map[int64][]string),
	}
}

// Начать новый список
func (b *ShoppingBot) startNewList(chatID int64) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.shoppingLists[chatID] = []string{} // Новый список для конкретного чата
	return "<b>🆕 Новый список покупок начат.</b> Просто отправляйте сообщения с пунктами списка!"
}

// Добавить пункты в список
func (b *ShoppingBot) addToList(chatID int64, items []string) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			b.shoppingLists[chatID] = append(b.shoppingLists[chatID], trimmed)
		}
	}
	return "<b>✅ Пункты добавлены в список.</b>"
}

// Получить текущий список
func (b *ShoppingBot) getList(chatID int64) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "<b>📋 Список покупок пуст.</b>"
	}

	var result strings.Builder
	result.WriteString("<b>📋 Ваш список покупок:</b>\n")
	for i, item := range list {
		// Формируем нумерованный список
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
	return result.String()
}

// Зачеркнуть пункт
func (b *ShoppingBot) strikeThrough(chatID int64, index int) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "", fmt.Errorf("<b>⚠️ Список пуст. Нечего вычеркивать.</b>")
	}

	if index < 1 || index > len(list) {
		return "", fmt.Errorf("<b>⚠️ Неверный номер. Пожалуйста, выберите существующий пункт.</b>")
	}

	// Зачеркнуть пункт
	item := list[index-1]
	list[index-1] = fmt.Sprintf("<s>%s</s>", item) // Используем HTML <s> для зачёркивания
	b.shoppingLists[chatID] = list
	return "<b>✅ Пункт вычеркнут.</b>", nil
}

// Отменить зачёркивание
func (b *ShoppingBot) unstrike(chatID int64, index int) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "", fmt.Errorf("<b>⚠️ Список пуст. Нечего отменять.</b>")
	}

	if index < 1 || index > len(list) {
		return "", fmt.Errorf("<b>⚠️ Неверный номер. Пожалуйста, выберите существующий пункт.</b>")
	}

	// Убираем зачёркивание, если оно есть
	item := list[index-1]
	if strings.HasPrefix(item, "<s>") && strings.HasSuffix(item, "</s>") {
		// Извлекаем оригинальный текст
		list[index-1] = item[3 : len(item)-4]
		b.shoppingLists[chatID] = list
		return "<b>✅ Зачёркивание отменено.</b>", nil
	}

	return "<b>⚠️ Этот пункт не был зачёркнут.</b>", nil
}

// Удалить пункт
func (b *ShoppingBot) deleteItem(chatID int64, index int) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "", fmt.Errorf("<b>⚠️ Список пуст. Нечего удалять.</b>")
	}

	if index < 1 || index > len(list) {
		return "", fmt.Errorf("<b>⚠️ Неверный номер. Пожалуйста, выберите существующий пункт.</b>")
	}

	// Удаляем пункт
	list = append(list[:index-1], list[index:]...) // Исключаем выбранный элемент
	b.shoppingLists[chatID] = list
	return "<b>✅ Пункт удалён.</b>", nil
}

func main() {
	// Загружаем переменные из .env файла
	err := godotenv.Load()
	if err := godotenv.Load(); err != nil {
    log.Printf("Warning: .env file not found, using environment variables")
}

	// Получаем токен из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatalf("TELEGRAM_BOT_TOKEN not set in environment")
	}

	log.Printf("Starting bot with token: %s", token)

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	// Удаляем Webhook, если он активен
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Fatalf("Failed to delete webhook: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	shoppingBot := NewShoppingBot()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID // Получаем chat ID
		var response string

		// Проверяем текст на команды
		switch {
		case strings.HasPrefix(update.Message.Text, "/start"):
			response = "👋 Привет! Я бот для списка покупок. Команды:\n" +
				"1️⃣ /newlist - начать новый список\n" +
				"2️⃣ /showlist - показать текущий список\n" +
				"3️⃣ /strike [№] - вычеркнуть пункт\n" +
				"4️⃣ /unstrike [№] - отменить вычеркивание\n" +
				"5️⃣ /delete [№] - удалить пункт\n"

		case strings.HasPrefix(update.Message.Text, "/newlist"):
			response = shoppingBot.startNewList(chatID)

		case strings.HasPrefix(update.Message.Text, "/showlist"):
			response = shoppingBot.getList(chatID)

		case strings.HasPrefix(update.Message.Text, "/strike"):
			arg := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/strike"))
			index, err := strconv.Atoi(arg)
			if err != nil {
				response = "<b>⚠️ Укажите корректный номер пункта для вычеркивания.</b>"
				log.Printf("Failed to parse strike index: %v", err)
			} else {
				list, err := shoppingBot.strikeThrough(chatID, index)
				if err != nil {
					response = err.Error()
				} else {
					response = list
				}
			}

		case strings.HasPrefix(update.Message.Text, "/unstrike"):
			arg := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/unstrike"))
			index, err := strconv.Atoi(arg)
			if err != nil {
				response = "<b>⚠️ Укажите корректный номер пункта для отмены зачёркивания.</b>"
				log.Printf("Failed to parse unstrike index: %v", err)
			} else {
				list, err := shoppingBot.unstrike(chatID, index)
				if err != nil {
					response = err.Error()
				} else {
					response = list
				}
			}

		case strings.HasPrefix(update.Message.Text, "/delete"):
			arg := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/delete"))
			index, err := strconv.Atoi(arg)
			if err != nil {
				response = "<b>⚠️ Укажите корректный номер пункта для удаления.</b>"
				log.Printf("Failed to parse delete index: %v", err)
			} else {
				list, err := shoppingBot.deleteItem(chatID, index)
				if err != nil {
					response = err.Error()
				} else {
					response = list
				}
			}

		default:
			// Если не команда, разбиваем сообщение на строки и добавляем в список
			lines := strings.Split(update.Message.Text, "\n")
			response = shoppingBot.addToList(chatID, lines)
		}

		// Отправляем ответ
		msg := tgbotapi.NewMessage(chatID, response)
		msg.ParseMode = "HTML" // Указываем HTML-разметку
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}
