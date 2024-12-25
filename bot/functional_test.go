package bot_test

import (
	"telegram-shopping-bot/bot"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartNewList(t *testing.T) {
	b := bot.NewShoppingBot()

	// Создание нового списка
	response := b.StartNewList(12345)
	assert.Equal(t, "<b>🆕 Новый список покупок начат.</b> Просто отправляйте сообщения с пунктами списка!", response)
}

func TestAddToList(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	// Добавление элементов
	response := b.AddToList(12345, []string{"Молоко", "Хлеб"})
	assert.Equal(t, "<b>✅ Пункты добавлены в список.</b>", response)

	// Проверяем список через GetList (с отформатированным выводом)
	expected := "<b>📋 Ваш список покупок:</b>\n1. Молоко\n2. Хлеб\n"
	actual := b.GetList(12345)
	assert.Equal(t, expected, actual, "Список должен быть корректно отформатирован")
}

func TestDeleteItem(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)
	b.AddToList(12345, []string{"Молоко", "Хлеб"})

	// Удаляем элемент
	response, err := b.DeleteItem(12345, 2) // Удаляем "Хлеб"
	assert.NoError(t, err)
	assert.Equal(t, "<b>✅ Пункт удалён.</b>", response)

	// Проверяем обновлённый список
	expected := "<b>📋 Ваш список покупок:</b>\n1. Молоко\n"
	actual := b.GetList(12345)
	assert.Equal(t, expected, actual, "Список должен содержать только 'Молоко'")
}

func TestStrikeThrough(t *testing.T) {
	b := bot.NewShoppingBot()
	chatID := int64(12345)

	b.StartNewList(chatID)
	b.AddToList(chatID, []string{"Молоко", "Хлеб", "Яблоки"})

	// Вычеркиваем второй пункт (Хлеб)
	resp, err := b.StrikeThrough(chatID, 2)
	assert.NoError(t, err)
	assert.Equal(t, "<b>✅ Пункт вычеркнут.</b>", resp)

	// Проверяем список
	list := b.GetList(chatID)
	// Ожидаем, что пункт 2 будет "<s>Хлеб</s>"
	expected := "<b>📋 Ваш список покупок:</b>\n1. Молоко\n2. <s>Хлеб</s>\n3. Яблоки\n"
	assert.Equal(t, expected, list, "Должны увидеть зачёркнутый элемент")
}