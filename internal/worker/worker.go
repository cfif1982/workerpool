package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"training/internal/task"
)

const taskTimeOut = time.Duration(4) * time.Second // таймаут выполнения одной задачи

type Worker struct {
	ID          int
	queueCH     chan *task.Task // канал с очередью задач
	closeWorker bool            // флаг закрытия воркера
}

func NewWorker(id int, queue chan *task.Task) *Worker {
	return &Worker{
		ID:          id,
		queueCH:     queue,
		closeWorker: false,
	}
}

// начало работы воркера
func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// если воркер помечен на закрытие, то завершаем задачу, а новую не берем и закрываемся
		if w.closeWorker {
			return
		}

		select {
		// берем задачу из очереди
		case t, ok := <-w.queueCH:
			// если канал закрыт
			if !ok {
				fmt.Printf("worker %d закрыт \n", w.ID)
				return
			}

			// задаем контекст для отмены задачи
			ctxTaskTimeOut, taskCancelTimeOut := context.WithTimeout(ctx, taskTimeOut)
			defer taskCancelTimeOut()

			fmt.Printf("worker %d начал задачу %d \n", w.ID, t.ID)

			// выполняем задачу
			t.Do(ctxTaskTimeOut)

			// проверяем, как завершился контекст
			switch ctxTaskTimeOut.Err() {
			case context.Canceled:
				fmt.Printf("Прервали работу задачи %d\n", t.ID)
			case context.DeadlineExceeded:
				fmt.Printf("Истекло время работы задачи %d\n", t.ID)
			default:
				fmt.Printf("worker %d закончил задачу %d\n", w.ID, t.ID)
			}

		// следим за контекстом для отмены
		case <-ctx.Done():
			fmt.Printf("worker %d отменен\n", w.ID)
			return
		}

	}
}

// помечаем  воркера на закрытие
func (w *Worker) Close() {
	w.closeWorker = true
}
