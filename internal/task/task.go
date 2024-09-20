package task

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cfif1982/workerpool/internal/result"
)

type Task struct {
	ID  int
	URL string
}

// конструктор
func NewTask(id int, url string) *Task {
	return &Task{
		ID:  id,
		URL: url,
	}
}

// выполнение задачи
func (t *Task) Do(ctx context.Context) *result.Result {
	fmt.Println("task started:", t.ID)

	// Создаем HTTP-запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", t.URL, nil)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return nil
	}

	// Засекаем время перед отправкой запроса
	startTime := time.Now()

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return nil
	}

	defer resp.Body.Close()

	// Засекаем время после получения ответа
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Читаем и выводим ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return nil
	}

	fmt.Println("task finished:", t.ID)

	// возвращаем результат работы
	return result.NewResult(body, duration)
}
