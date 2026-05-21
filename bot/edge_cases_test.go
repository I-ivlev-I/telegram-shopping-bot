package bot_test

import (
	"telegram-shopping-bot/bot"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestAddInvalidItems(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	response := b.AddToList(12345, []string{"", "   "})
	assert.Equal(t, "<b>✅ Пункты добавлены в список.</b>", response)

	assert.Equal(t, "<b>📋 Список покупок пуст.</b>", b.GetList(12345))
}

func TestDeleteInvalidIndexes(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)
	b.AddToList(12345, []string{"Молоко"})

	_, err := b.DeleteItem(12345, -1)
	assert.Error(t, err)

	_, err = b.DeleteItem(12345, 2)
	assert.Error(t, err)
}

func TestEmptyListOperations(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	_, err := b.DeleteItem(12345, 1)
	assert.Error(t, err)

	_, err = b.StrikeThrough(12345, 1)
	assert.Error(t, err)
}

func TestHandleCallbackAndKeyboard(t *testing.T) {
	b := bot.NewShoppingBot()
	chatID := int64(12345)
	b.StartNewList(chatID)
	b.AddToList(chatID, []string{"Хлеб", "Молоко"})

	resp := bot.HandleCallback(b, "str:1", chatID)
	assert.Equal(t, "<b>✅ Пункт вычеркнут.</b>", resp)

	kb := b.BuildListKeyboard(chatID)
	if assert.NotNil(t, kb) {
		assert.Len(t, kb.InlineKeyboard, 2)
		if assert.NotNil(t, kb.InlineKeyboard[0][1].CallbackData) {
			assert.Equal(t, "uns:1", *kb.InlineKeyboard[0][1].CallbackData)
		}
	}

	resp = bot.HandleCallback(b, "bad:data", chatID)
	assert.Contains(t, resp, "Не удалось обработать")
}

func TestStrictCommandParsingWithEntities(t *testing.T) {
	b := bot.NewShoppingBot()
	msg := &tgbotapi.Message{
		Text: "/unknown",
		Chat: &tgbotapi.Chat{ID: 12345},
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 8},
		},
	}
	response := bot.HandleUpdate(b, msg)
	assert.Contains(t, response, "Неизвестная команда")
}
