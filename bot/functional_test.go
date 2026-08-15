package bot_test

import (
	"telegram-shopping-bot/bot"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestStartNewList(t *testing.T) {
	b := bot.NewShoppingBot()

	response := b.StartNewList(12345)
	assert.Equal(t, "<b>🆕 Новый список покупок начат.</b> Просто отправляйте сообщения с пунктами списка!", response)
}

func TestAddToList(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	response := b.AddToList(12345, []string{"Молоко", "Хлеб"})
	assert.Equal(t, "<b>✅ Пункты добавлены в список.</b>", response)

	expected := "<b>📋 Ваш список покупок:</b>\n1. Молоко\n2. Хлеб\n"
	actual := b.GetList(12345)
	assert.Equal(t, expected, actual, "Список должен быть корректно отформатирован")
}

func TestDeleteItem(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)
	b.AddToList(12345, []string{"Молоко", "Хлеб"})

	response, err := b.DeleteItem(12345, 2)
	assert.NoError(t, err)
	assert.Equal(t, "<b>✅ Пункт удалён.</b>", response)

	expected := "<b>📋 Ваш список покупок:</b>\n1. Молоко\n"
	actual := b.GetList(12345)
	assert.Equal(t, expected, actual, "Список должен содержать только 'Молоко'")
}

func TestStrikeThrough(t *testing.T) {
	b := bot.NewShoppingBot()
	chatID := int64(12345)

	b.StartNewList(chatID)
	b.AddToList(chatID, []string{"Молоко", "Хлеб", "Яблоки"})

	resp, err := b.StrikeThrough(chatID, 2)
	assert.NoError(t, err)
	assert.Equal(t, "<b>✅ Пункт вычеркнут.</b>", resp)

	list := b.GetList(chatID)
	expected := "<b>📋 Ваш список покупок:</b>\n1. Молоко\n2. <s>Хлеб</s>\n3. Яблоки\n"
	assert.Equal(t, expected, list, "Должны увидеть зачёркнутый элемент")
}

func TestUnstrike(t *testing.T) {
	b := bot.NewShoppingBot()
	chatID := int64(12345)

	b.StartNewList(chatID)
	b.AddToList(chatID, []string{"Молоко", "Хлеб", "Яблоки"})

	_, err := b.StrikeThrough(chatID, 2)
	assert.NoError(t, err)

	resp, err := b.Unstrike(chatID, 2)
	assert.NoError(t, err)
	assert.Equal(t, "<b>✅ Зачёркивание отменено.</b>", resp)

	list := b.GetList(chatID)
	expected := "<b>📋 Ваш список покупок:</b>\n1. Молоко\n2. Хлеб\n3. Яблоки\n"
	assert.Equal(t, expected, list, "После /unstrike элемент 2 должен быть обычным текстом")
}

func TestEscapesHTMLInUserInput(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)
	b.AddToList(12345, []string{"<b>milk</b>", "<a href='x'>x</a>"})

	list := b.GetList(12345)
	assert.Contains(t, list, "&lt;b&gt;milk&lt;/b&gt;")
	assert.Contains(t, list, "&lt;a href=&#39;x&#39;&gt;x&lt;/a&gt;")
}

func TestHandleUpdateCommandParsing(t *testing.T) {
	b := bot.NewShoppingBot()
	msg := &tgbotapi.Message{
		Text:     "/start123",
		Chat:     &tgbotapi.Chat{ID: 12345},
		Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 9}},
	}

	response := bot.HandleUpdate(b, msg)
	assert.Equal(t, "<b>⚠️ Неизвестная команда.</b> Используйте /start, чтобы посмотреть список команд.", response)
}
