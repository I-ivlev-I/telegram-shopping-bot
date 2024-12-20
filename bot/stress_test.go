package bot_test

import (
	"fmt"
	"sync"
	"telegram-shopping-bot/bot"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStressAddLargeNumberOfItems(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	// Добавляем 10,000 элементов
	for i := 0; i < 10000; i++ {
		b.AddToList(12345, []string{fmt.Sprintf("Item %d", i+1)})
	}

	// Проверяем количество элементов через GetList
	list := b.GetList(12345)
	assert.Contains(t, list, "1. Item 1")
	assert.Contains(t, list, "10000. Item 10000")
}

func TestStressDeleteLargeNumberOfItems(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	// Добавляем 10,000 элементов
	for i := 0; i < 10000; i++ {
		b.AddToList(12345, []string{fmt.Sprintf("Item %d", i+1)})
	}

	// Удаляем элементы
	for i := 10000; i > 0; i-- {
		_, err := b.DeleteItem(12345, 1)
		assert.NoError(t, err)
	}

	// Проверяем, что список пуст
	assert.Equal(t, "<b>📋 Список покупок пуст.</b>", b.GetList(12345))
}

func TestStressConcurrentAccess(t *testing.T) {
	b := bot.NewShoppingBot()
	b.StartNewList(12345)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			b.AddToList(12345, []string{fmt.Sprintf("Item %d", index+1)})
		}(i)

		wg.Add(1)
		go func() {
			defer wg.Done()
			b.DeleteItem(12345, 1)
		}()
	}

	wg.Wait()
	assert.Contains(t, b.GetList(12345), "<b>📋 Ваш список покупок:</b>")
}
