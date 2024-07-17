package task

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

const taskMaxDuration = 10 // максимальное время для rand

type Task struct {
	ID       int
	duration time.Duration // время выполнения одной задачи
}

func NewTask(id int) *Task {
	return &Task{
		ID:       id,
		duration: time.Duration((rand.Intn(taskMaxDuration-1) + 1)) * time.Second, // случайное время на выполнение задачи,
	}
}

// выполнение задачи
func (t *Task) Do(ctx context.Context) {
	deadline, _ := ctx.Deadline() // узнаем время для отмены задачи, чтобы вывести эти данные
	sleepTime := time.Until(deadline)

	fmt.Printf("задача %d начата, продолжительность: %f, deadline: %f\n", t.ID, t.duration.Seconds(), sleepTime.Seconds())

	for {
		select {
		// следим за контекстом для отмены работы
		case <-ctx.Done():
			fmt.Printf("задача %d отменена\n", t.ID)
			return
		// выводим сообщение о завершении работы
		case <-time.After(t.duration):
			fmt.Printf("задача %d закончена\n", t.ID)
			return
		}
	}

}
