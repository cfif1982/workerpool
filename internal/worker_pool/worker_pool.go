package workerpool

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"training/internal/task"
	"training/internal/worker"
)

type TaskReceiverI interface {
	GetTask() (string, bool)
}

type ResultSaverI interface {
	SaveResult(str string)
}

type WorkerPool struct {
	countTasks       int
	countWorkers     int
	totalWorkTimeOut time.Duration
	queueCH          chan *task.Task
	wg               sync.WaitGroup
	TaskReceiverI
	ResultSaverI
}

func NewWorkerPool(countTasks, countWorkers, intTotalWorkTimeOut int, tr TaskReceiverI, rs ResultSaverI) *WorkerPool {
	return &WorkerPool{
		countTasks:       countTasks,
		countWorkers:     countWorkers,
		totalWorkTimeOut: time.Duration(intTotalWorkTimeOut) * time.Second,
		queueCH:          make(chan *task.Task), // канал для передачи задач,
		TaskReceiverI:    tr,
		ResultSaverI:     rs,
	}
}

// генерируем задачи в очередь
func (p *WorkerPool) generateTask(wg *sync.WaitGroup) {
	defer wg.Done()

	// создаем задачи и помещаем их в очередь
	for i := 1; i <= p.countTasks; i++ {
		t := task.NewTask(i)
		p.queueCH <- t
	}

	// закрывает канал тот , кто в него пишет
	close(p.queueCH)
}

func (p *WorkerPool) Start() {

	var wg sync.WaitGroup

	// Обработка сигналов завершения работы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// создаем контекст для отены общего выполнения всех задач
	ctxTotalTimeOut, totalTimeOutCancel := context.WithTimeout(context.Background(), p.totalWorkTimeOut)
	defer totalTimeOutCancel()

	// созадем воркеров
	// мы можем здесь вернуть указатели на воркеров и хранить их в слайсе.
	// И в нужный момент обращаться к ним и закрывать, вызвав метод close()
	// функцию addWorker мы можем вызывать в любой момент и добавлять новых воркеров
	for i := 1; i <= p.countWorkers; i++ {
		addWorker(ctxTotalTimeOut, &wg, i, p.queueCH)
	}

	wg.Add(1)

	// генерируем задачи
	go p.generateTask(&wg)

	// Дополнительный канал для уведомления о завершении всех worker'ов
	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	// следим за сигнальным каналом
	case <-sigChan:
		fmt.Println("Grasefull Shutdown start")
		totalTimeOutCancel()
	// следим за каналом окончания работы
	case <-doneChan:
		fmt.Println("все воркеры завершили работу")
	// следим за контекстом отмены всех работ
	case <-ctxTotalTimeOut.Done():
		fmt.Printf("время работы истекло\n")
	}

	// проверяем, как завершился контекст
	switch ctxTotalTimeOut.Err() {
	case context.Canceled:
		fmt.Println("Прервали общую работу")
	case context.DeadlineExceeded:
		fmt.Println("Истекло время общей работы")
	}

	fmt.Println("ФИНИШ")
}

func addWorker(ctx context.Context, wg *sync.WaitGroup, id int, queueCH chan *task.Task) {
	// созадем воркера
	worker := worker.NewWorker(id, queueCH)

	wg.Add(1)

	// запускаем воркера в работу
	go worker.Start(ctx, wg)
}
