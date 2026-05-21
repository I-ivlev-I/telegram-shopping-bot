package bot

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	BtnNewList  = "🆕 New List"
	BtnShowList = "📋 Show List"
	BtnDelete   = "🗑 Delete"
	BtnStrike   = "✅ Strike"
	BtnUnstrike = "↩️ Unstrike"
	BtnHelp     = "ℹ️ Help"
)

type ShoppingBot struct {
	mu            sync.Mutex
	shoppingLists map[int64][]string
}

func NewShoppingBot() *ShoppingBot {
	return &ShoppingBot{shoppingLists: make(map[int64][]string)}
}

func MainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnNewList),
			tgbotapi.NewKeyboardButton(BtnShowList),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnDelete),
			tgbotapi.NewKeyboardButton(BtnStrike),
			tgbotapi.NewKeyboardButton(BtnUnstrike),
		),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(BtnHelp)),
	)
}

func startText() string {
	return "👋 Привет! Я бот для списка покупок. Команды:\n" +
		"/newlist - начать новый список\n" +
		"/showlist - показать список и кнопки действий\n" +
		"/delete [№] - удалить пункт\n" +
		"/strike [№] - вычеркнуть пункт\n" +
		"/unstrike [№] - отменить зачёркивание\n\n" +
		"Или используйте кнопки меню ниже."
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
			b.shoppingLists[chatID] = append(b.shoppingLists[chatID], html.EscapeString(trimmed))
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
func isStruck(item string) bool {
	return strings.HasPrefix(item, "<s>") && strings.HasSuffix(item, "</s>")
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
	if isStruck(list[index-1]) {
		return "<b>⚠️ Этот пункт уже зачёркнут.</b>", nil
	}
	list[index-1] = "<s>" + list[index-1] + "</s>"
	b.shoppingLists[chatID] = list
	return "<b>✅ Пункт вычеркнут.</b>", nil
}
func (b *ShoppingBot) Unstrike(chatID int64, index int) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	list, exists := b.shoppingLists[chatID]
	if !exists || len(list) == 0 {
		return "", fmt.Errorf("<b>⚠️ Список пуст. Нечего отменять.</b>")
	}
	if index < 1 || index > len(list) {
		return "", fmt.Errorf("<b>⚠️ Неверный номер. Пожалуйста, выберите существующий пункт.</b>")
	}
	item := list[index-1]
	if isStruck(item) {
		list[index-1] = item[3 : len(item)-4]
		b.shoppingLists[chatID] = list
		return "<b>✅ Зачёркивание отменено.</b>", nil
	}
	return "<b>⚠️ Этот пункт не был зачёркнут.</b>", nil
}

func (b *ShoppingBot) buildListButtons(chatID int64) [][]tgbotapi.InlineKeyboardButton {
	b.mu.Lock()
	defer b.mu.Unlock()
	list := b.shoppingLists[chatID]
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(list)+1)
	for i, item := range list {
		idx := i + 1
		action := "str"
		label := fmt.Sprintf("✅ Вычеркнуть %d", idx)
		if isStruck(item) {
			action = "uns"
			label = fmt.Sprintf("↩️ Отменить %d", idx)
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🗑 Удалить %d", idx), fmt.Sprintf("del:%d", idx)), tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%d", action, idx))))
	}
	return rows
}
func (b *ShoppingBot) BuildListKeyboard(chatID int64) *tgbotapi.InlineKeyboardMarkup {
	rows := b.buildListButtons(chatID)
	if len(rows) == 0 {
		return nil
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &markup
}

func handleNumberAction(b *ShoppingBot, chatID int64, prefix, arg string) string {
	index, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil {
		switch prefix {
		case "delete":
			return "<b>⚠️ Укажите корректный номер пункта для удаления.</b>"
		case "strike":
			return "<b>⚠️ Укажите корректный номер пункта для вычеркивания.</b>"
		default:
			return "<b>⚠️ Укажите корректный номер пункта для отмены зачёркивания.</b>"
		}
	}
	var response string
	if prefix == "delete" {
		response, err = b.DeleteItem(chatID, index)
	} else if prefix == "strike" {
		response, err = b.StrikeThrough(chatID, index)
	} else {
		response, err = b.Unstrike(chatID, index)
	}
	if err != nil {
		return err.Error()
	}
	return response
}

func HandleUpdate(b *ShoppingBot, message *tgbotapi.Message) string {
	chatID := message.Chat.ID
	text := strings.TrimSpace(message.Text)

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			return startText()
		case "newlist":
			return b.StartNewList(chatID)
		case "showlist":
			return b.GetList(chatID)
		case "delete", "strike", "unstrike":
			return handleNumberAction(b, chatID, message.Command(), message.CommandArguments())
		default:
			return "<b>⚠️ Неизвестная команда.</b> Используйте /start, чтобы посмотреть список команд."
		}
	}

	switch text {
	case BtnHelp:
		return startText()
	case BtnNewList:
		return b.StartNewList(chatID)
	case BtnShowList:
		return b.GetList(chatID)
	case BtnDelete:
		return "<b>🗑 Укажите номер:</b> отправьте команду в формате <code>/delete 2</code>."
	case BtnStrike:
		return "<b>✅ Укажите номер:</b> отправьте команду в формате <code>/strike 2</code>."
	case BtnUnstrike:
		return "<b>↩️ Укажите номер:</b> отправьте команду в формате <code>/unstrike 2</code>."
	default:
		return b.AddToList(chatID, strings.Split(text, "\n"))
	}
}

func HandleCallback(b *ShoppingBot, callbackData string, chatID int64) string {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 2 {
		return "<b>⚠️ Не удалось обработать действие.</b>"
	}
	index, err := strconv.Atoi(parts[1])
	if err != nil {
		return "<b>⚠️ Не удалось обработать действие.</b>"
	}
	switch parts[0] {
	case "del":
		resp, err := b.DeleteItem(chatID, index)
		if err != nil {
			return err.Error()
		}
		return resp
	case "str":
		resp, err := b.StrikeThrough(chatID, index)
		if err != nil {
			return err.Error()
		}
		return resp
	case "uns":
		resp, err := b.Unstrike(chatID, index)
		if err != nil {
			return err.Error()
		}
		return resp
	default:
		return "<b>⚠️ Неизвестное действие.</b>"
	}
}
