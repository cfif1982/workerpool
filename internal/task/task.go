package task

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cfif1982/workerpool/internal/result"
)

type ResultSaverI interface {
	SaveResult(result *result.Result)
}

type Task struct {
	ID  int
	URL string
	rs  ResultSaverI
}

func NewTask(id int, url string, rs ResultSaverI) *Task {
	return &Task{
		ID:  id,
		URL: url,
		rs:  rs,
	}
}

// выполнение задачи
func (t *Task) Do(ctx context.Context) {

	// Создаем HTTP-запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", t.URL, nil)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	// Засекаем время перед отправкой запроса
	startTime := time.Now()

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return
	}
	defer resp.Body.Close()

	// Засекаем время после получения ответа
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Читаем и выводим ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return
	}

	result := result.NewResult(body, duration)

	t.rs.SaveResult(result)
}
