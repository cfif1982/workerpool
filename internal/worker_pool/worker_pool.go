package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/cfif1982/workerpool/internal/result"
	"github.com/cfif1982/workerpool/internal/task"
	"github.com/cfif1982/workerpool/internal/worker"
)

const addWorkersRequestDuration = time.Duration(2) * time.Second    // максимальное время выполнения запроса для увеличения количества воркеров
const deleteWorkersRequestDuration = time.Duration(1) * time.Second // минимальное время выполнения запроса для уменьшения количества воркеров
const changeWorkersCountTime = time.Duration(5) * time.Second       // время для таймера изменения количества воркеров

// интерфейс task receiver. Задачи можем получать из файла, из БД, или вообще из интернета
type TaskReceiverI interface {
	Start(ctx context.Context, wg *sync.WaitGroup) // запуск task receiver
	GetTaskCH() chan *task.Task                    // возвращает канал, по которому  будeт отсылаться задачи
}

// интерфейс result saver. Результат может сохраняться в канал, в БД или еще куда
type ResultSaverI interface {
	SaveResult(result *result.Result, ok bool) // сохраняет результа
}

type WorkerPool struct {
	countWorkers       int
	workerResultCH     chan *result.Result // канал для результатов воркеров
	wg                 sync.WaitGroup
	tr                 TaskReceiverI
	rs                 ResultSaverI
	workers            []*worker.Worker // слайс для хранения воркеров. Нужен для доступа к ним когда нужно закрыть
	averageRequestTime time.Duration    // среднее время выполнения запроса
}

// конструктор
func NewWorkerPool(countWorkers int, tr TaskReceiverI, rs ResultSaverI) *WorkerPool {
	return &WorkerPool{
		countWorkers: countWorkers,
		// TODO: нужно ли его делать буферизированным?
		workerResultCH: make(chan *result.Result), // канал для передачи результата от воркеров
		tr:             tr,
		rs:             rs,
		workers:        make([]*worker.Worker, 0),
	}
}

// запуск worker pool
// возвращаем сигнальный канал о завершении работы  worker pool
func (p *WorkerPool) Start(ctx context.Context) chan struct{} {

	fmt.Println("СТАРТ")

	// Засекаем время перед началом работы
	startTime := time.Now()

	// созадем воркеров
	for i := 1; i <= p.countWorkers; i++ {
		// передаем канал для получения задач и канал для возврата результат воркера
		p.addWorker(ctx, i, p.tr.GetTaskCH(), p.workerResultCH)
	}

	p.wg.Add(1) // увеличиваем wg для того, чтобы сгенерировать задачи и не закрытьтся раньше времени

	// начинаем генерировать задачи
	p.tr.Start(ctx, &p.wg)

	// канал для уведомления о завершении всех worker'ов
	doneChan := make(chan struct{})

	// запускаем горутину, которая ждет завершения всех воркеров и закрывает канал
	// TODO: доделать gracefull shutdown. Нужно тут его делать или эта горутина сама закроется при завершении программы?
	go func() {
		p.wg.Wait()
		close(doneChan)
	}()

	// запускаем сохранение результата от воркеров в Result Saver
	go p.saveResults(ctx)

	// запускаем таймер изменения количества воркеров
	go p.changeWorkersCountTimer(ctx)

	select {
	// следим за каналом окончания работы
	case <-doneChan:
		fmt.Println("все воркеры завершили работу")
	}

	// Засекаем время после получения ответа
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Println("*************************************************")
	fmt.Printf("общее время выполнения работы: %f s\n", float64(duration)/float64(time.Second))
	fmt.Printf("воркеров: %d, логических ядер: %d, логических процессоров: %d\n", p.countWorkers, runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("среднее время выполнения запроса: %f ms\n", float64(p.averageRequestTime)/float64(time.Millisecond))
	fmt.Println("ФИНИШ")

	// возвращаем сигнальный канал о завершении работы  worker pool
	return doneChan
}

// добавляем воркера
// taskCH - канал, по которому получаем задачи
// resultCH - канал, по которому отправляем результаты работы
func (p *WorkerPool) addWorker(ctx context.Context, ind int, taskCH chan *task.Task, resultCH chan *result.Result) {

	// созадем воркера
	worker := worker.NewWorker(ind, taskCH, resultCH)

	// добавляем воркера в слайс
	p.workers = append(p.workers, worker)

	// дожидаемся окончания его работы
	p.wg.Add(1)

	// запускаем воркера в работу
	go worker.Start(ctx, &p.wg)
}

// сохраняем результат в Result Saver
func (p *WorkerPool) saveResults(ctx context.Context) {

	for {
		select {
		// следим за закрытием контекста
		case <-ctx.Done():
			return

		// берем результат из очереди
		case result, ok := <-p.workerResultCH:

			// сохраняем результат
			p.rs.SaveResult(result, ok)

			// считаем среднее время всех запросов
			if p.averageRequestTime == 0 {
				p.averageRequestTime = result.Time
			} else {
				p.averageRequestTime = (p.averageRequestTime + result.Time) / 2
			}
		}
	}
}

// таймер для изменения количества воркеров
func (p *WorkerPool) changeWorkersCountTimer(ctx context.Context) {

	ticker := time.NewTicker(changeWorkersCountTime * time.Second)
	defer ticker.Stop()

	for {
		select {
		// следим за закрытием контекста
		case <-ctx.Done():
			return

		case <-ticker.C:
			// в зависимости от времени добавляем воркеров
			if p.averageRequestTime > addWorkersRequestDuration {
				p.addWorker(ctx, len(p.workers), p.tr.GetTaskCH(), p.workerResultCH)
			}

			// в зависимости от времени убираем воркеров
			if p.averageRequestTime < deleteWorkersRequestDuration {
				// находим индекс последнего воркера
				lastInd := len(p.workers) - 1

				// закрываем последнего воркера
				p.workers[lastInd].Close()

				// удаляем воркера из слайса
				p.workers = p.workers[:lastInd]
			}
		}
	}
}
