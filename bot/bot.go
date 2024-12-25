package bot

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ShoppingBot struct {
	mu            sync.Mutex
	shoppingLists map[int64][]string
}

func NewShoppingBot() *ShoppingBot {
	return &ShoppingBot{
		shoppingLists: make(map[int64][]string),
	}
}

func (b *ShoppingBot) StartNewList(chatID int64) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.shoppingLists[chatID] = []string{}
	return "<b>🆕 Новый список покупок начат.</b> Просто отправляйте сообщения с пунктами списка!"
}

func (b *ShoppingBot) AddToList(chatID int64, items []string) string {
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

func (b *ShoppingBot) GetList(chatID int64) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "<b>📋 Список покупок пуст.</b>"
	}

	var result strings.Builder
	result.WriteString("<b>📋 Ваш список покупок:</b>\n")
	for i, item := range list {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
	return result.String()
}

func (b *ShoppingBot) DeleteItem(chatID int64, index int) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "", fmt.Errorf("<b>⚠️ Список пуст. Нечего удалять.</b>")
	}
	if index < 1 || index > len(list) {
		return "", fmt.Errorf("<b>⚠️ Неверный номер. Пожалуйста, выберите существующий пункт.</b>")
	}
	b.shoppingLists[chatID] = append(list[:index-1], list[index:]...)
	return "<b>✅ Пункт удалён.</b>", nil
}

func (b *ShoppingBot) StrikeThrough(chatID int64, index int) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "", fmt.Errorf("<b>⚠️ Список пуст. Нечего вычеркивать.</b>")
	}
	if index < 1 || index > len(list) {
		return "", fmt.Errorf("<b>⚠️ Неверный номер. Пожалуйста, выберите существующий пункт.</b>")
	}
	list[index-1] = "<s>" + list[index-1] + "</s>"
	b.shoppingLists[chatID] = list
	return "<b>✅ Пункт вычеркнут.</b>", nil
}

func HandleUpdate(b *ShoppingBot, message *tgbotapi.Message) string {
	switch {
	case strings.HasPrefix(message.Text, "/start"):
		return "👋 Привет! Я бот для списка покупок. Команды:\n" +
			"/newlist - начать новый список\n" +
			"/showlist - показать список\n" +
			"/delete [№] - удалить пункт\n" +
			"/strike [№] - вычеркнуть пункт\n"

	case strings.HasPrefix(message.Text, "/newlist"):
		return b.StartNewList(message.Chat.ID)

	case strings.HasPrefix(message.Text, "/showlist"):
		return b.GetList(message.Chat.ID)

	case strings.HasPrefix(message.Text, "/delete"):
		arg := strings.TrimSpace(strings.TrimPrefix(message.Text, "/delete"))
		index, err := strconv.Atoi(arg)
		if err != nil {
			return "<b>⚠️ Укажите корректный номер пункта для удаления.</b>"
		}
		response, err := b.DeleteItem(message.Chat.ID, index)
		if err != nil {
			return err.Error()
		}
		return response

	case strings.HasPrefix(message.Text, "/strike"):
		arg := strings.TrimSpace(strings.TrimPrefix(message.Text, "/strike"))
		index, err := strconv.Atoi(arg)
		if err != nil {
			return "<b>⚠️ Укажите корректный номер пункта для вычеркивания.</b>"
		}
		response, err := b.StrikeThrough(message.Chat.ID, index)
		if err != nil {
			return err.Error()
		}
		return response

	default:
		// Если не команда - добавляем к списку
		lines := strings.Split(message.Text, "\n")
		return b.AddToList(message.Chat.ID, lines)
	}
}

