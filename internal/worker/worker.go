package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cfif1982/workerpool/internal/result"
	"github.com/cfif1982/workerpool/internal/task"
)

const taskTimeOut = time.Duration(2) * time.Second // таймаут выполнения одной задачи

type Worker struct {
	ID          int
	taskCH      chan *task.Task     // канал с очередью задач
	resultCH    chan *result.Result // канал куда отправляется результат
	closeWorker bool                // флаг закрытия воркера
}

// конструктор
func NewWorker(id int, taskCH chan *task.Task, resultCH chan *result.Result) *Worker {
	return &Worker{
		ID:          id,
		taskCH:      taskCH,
		resultCH:    resultCH,
		closeWorker: false,
	}
}

// начало работы воркера
func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done() // по окончании работы уменьшаем wg

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
		case t, ok := <-w.taskCH:
			// если канал закрыт
			if !ok {
				fmt.Printf("worker %d завершен, т.к. канал задач закрыт \n", w.ID)
				return
			}

			// задаем контекст для отмены задачи по времени
			ctxTaskTimeOut, taskCancel := context.WithTimeout(ctx, taskTimeOut)
			defer taskCancel()

			fmt.Printf("worker %d начал задачу %d \n", w.ID, t.ID)

			// выполняем задачу и отправялем рузельтат в канал
			w.resultCH <- t.Do(ctxTaskTimeOut)

			// проверяем, как завершился контекст
			switch ctxTaskTimeOut.Err() {
			case context.Canceled:
				fmt.Printf("Прервали работу задачи %d\n", t.ID)
			case context.DeadlineExceeded:
				fmt.Printf("Истекло время работы задачи %d\n", t.ID)
			default:
				fmt.Printf("worker %d закончил задачу %d\n", w.ID, t.ID)
			}
		}
	}
}

// помечаем  воркера на закрытие
func (w *Worker) Close() {
	w.closeWorker = true
}
