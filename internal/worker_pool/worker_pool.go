package workerpool

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cfif1982/workerpool/internal/result"
	"github.com/cfif1982/workerpool/internal/task"
	"github.com/cfif1982/workerpool/internal/worker"
)

type TaskReceiverI interface {
	Start(wg *sync.WaitGroup)
	GetTask(ctx context.Context) (*task.Task, bool) // передаем контекст для отмены ожидания задачи
}

type ResultSaverI interface {
	SaveResult(result *result.Result)
	GetResult() (*result.Result, bool)
}

type WorkerPool struct {
	countWorkers int
	queueCH      chan *task.Task
	wg           sync.WaitGroup
	TaskReceiverI
	ResultSaverI
}

func NewWorkerPool(countWorkers int, tr TaskReceiverI, rs ResultSaverI) *WorkerPool {
	return &WorkerPool{
		countWorkers:  countWorkers,
		queueCH:       make(chan *task.Task), // канал для передачи задач,
		TaskReceiverI: tr,
		ResultSaverI:  rs,
	}
}

func (p *WorkerPool) Start() {

	fmt.Println("СТАРТ")

	// Обработка сигналов завершения работы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// создаем контекст для отены общего выполнения всех задач
	ctxCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// созадем воркеров
	// мы можем здесь вернуть указатели на воркеров и хранить их в слайсе.
	// И в нужный момент обращаться к ним и закрывать, вызвав метод close()
	// функцию addWorker мы можем вызывать в любой момент и добавлять новых воркеров
	for i := 1; i <= p.countWorkers; i++ {
		p.addWorker(ctxCancel, i, p.queueCH)
	}

	p.wg.Add(1) // увеличиваем wg для того, чтобы сгенерировать задачи и не закрытьтся раньше времени

	// начинаем генерировать задачи
	p.TaskReceiverI.Start(&p.wg)

	// канал для уведомления о завершении всех worker'ов
	doneChan := make(chan struct{})

	// запускаем горутину, которая ждет завершения всех воркеров и закрывает канал
	// TODO: доделать gracefull shutdown. Нужно тут его делать или эта горутина сама закроется при завершении программы?
	go func() {
		p.wg.Wait()
		close(doneChan)
	}()

	// запускаем горутину, которая получает результат и регулирует количество воркеров
	// TODO: доделать gracefull shutdown. Нужно тут его делать или эта горутина сама закроется при завершении программы?
	go func() {
		for {
			result := p.GetResult()
		}
	}()

	select {
	// следим за сигнальным каналом
	case <-sigChan:
		fmt.Println("Grasefull Shutdown start")
		cancel() // отменяем работу всех воркеров при помощи контекста

	// следим за каналом окончания работы
	case <-doneChan:
		fmt.Println("все воркеры завершили работу")
	}

	// проверяем, как завершился контекст
	switch ctxCancel.Err() {
	case context.Canceled:
		fmt.Println("Прервали общую работу")
	}

	fmt.Println("ФИНИШ")
}

func (p *WorkerPool) addWorker(ctx context.Context, id int, queueCH chan *task.Task) {
	// созадем воркера
	worker := worker.NewWorker(id, queueCH)

	p.wg.Add(1)

	// запускаем воркера в работу
	go worker.Start(ctx, &p.wg, p.TaskReceiverI)
}
