package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cfif1982/workerpool/internal/task"
)

type TaskReceiverI interface {
	GetTask(ctx context.Context) (*task.Task, bool) // передаем контекст для отмены ожидания задачи
}

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
func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup, tr TaskReceiverI) {

	defer wg.Done()

	for {
		// если воркер помечен на закрытие, то завершаем задачу, а новую не берем и закрываемся
		if w.closeWorker {
			return
		}

		select {
		// следим за контекстом для отмены
		case <-ctx.Done():
			fmt.Printf("worker %d отменен\n", w.ID)
			return

		// берем задачу из очереди
		default:
			task, opened := tr.GetTask(ctx)

			// если канал закрыт
			if !opened {
				fmt.Printf("worker %d закрыт \n", w.ID)
				return
			}

			// задаем контекст для отмены задачи по времени
			ctxTaskTimeOut, taskCancel := context.WithTimeout(ctx, taskTimeOut)
			defer taskCancel()

			fmt.Printf("worker %d начал задачу %d \n", w.ID, task.ID)

			// выполняем задачу
			task.Do(ctxTaskTimeOut)

			// проверяем, как завершился контекст
			switch ctxTaskTimeOut.Err() {
			case context.Canceled:
				fmt.Printf("Прервали работу задачи %d\n", task.ID)
			case context.DeadlineExceeded:
				fmt.Printf("Истекло время работы задачи %d\n", task.ID)
			default:
				fmt.Printf("worker %d закончил задачу %d\n", w.ID, task.ID)
			}

		}

	}
}

// помечаем  воркера на закрытие
func (w *Worker) Close() {
	w.closeWorker = true
}
