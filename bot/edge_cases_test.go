package bot_test

import (
	"telegram-shopping-bot/bot"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddInvalidItems(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	// Добавление пустых строк
	response := b.AddToList(12345, []string{"", "   "})
	assert.Equal(t, "<b>✅ Пункты добавлены в список.</b>", response)

	// Проверяем, что список остаётся пустым
	assert.Equal(t, "<b>📋 Список покупок пуст.</b>", b.GetList(12345))
}

func TestDeleteInvalidIndexes(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)
	b.AddToList(12345, []string{"Молоко"})

	// Удаление с неправильным индексом
	_, err := b.DeleteItem(12345, -1)
	assert.Error(t, err)

	_, err = b.DeleteItem(12345, 2)
	assert.Error(t, err)
}

func TestEmptyListOperations(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	// Удаление из пустого списка
	_, err := b.DeleteItem(12345, 1)
	assert.Error(t, err)

	// Попытка зачеркнуть из пустого списка
	_, err = b.StrikeThrough(12345, 1)
	assert.Error(t, err)
}
