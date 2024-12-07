package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	rschan "github.com/cfif1982/workerpool/internal/result_saver/result_saver_chan"
	trfile "github.com/cfif1982/workerpool/internal/task_receiver/task_receiver_file"
	workerpool "github.com/cfif1982/workerpool/internal/worker_pool"
)

const countWorkers = 1 // начальное количество воркеров

func main() {
	// создаем task receiver
	tr := trfile.NewTaskReceiverFile("addr.txt")

	// создаем result saver
	rs := rschan.NewResultSaver()

	// создаем Worker Pool
	wp := workerpool.NewWorkerPool(countWorkers, tr, rs)

	// запускаем таймер для отображения счетчика времени
	go timer()

	// Обработка сигналов завершения работы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// создаем контекст для отены worker Pool
	ctxCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	// запускаем worker pool
	// Засекаем время перед началом работы
	startTime := time.Now()

	// возвращается сигнальный канал о завершении работы  worker pool
	wpDoneCh := wp.Start(ctxCancel)

	select {
	// следим за сигнальным каналом
	case <-sigChan:
		fmt.Println("Grasefull Shutdown start")
		cancel() // отменяем работу всех воркеров при помощи контекста

	// если работа worker pool завершена, то завершаем программу
	case <-wpDoneCh:
		fmt.Println("работа worker pool завершена")
		// Засекаем время после получения ответа
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		fmt.Println("*************************************************")
		fmt.Printf("общее время выполнения работы: %f s\n", float64(duration)/float64(time.Second))
		fmt.Printf("воркеров: %d, логических ядер: %d, логических процессоров: %d\n", countWorkers, runtime.NumCPU(), runtime.GOMAXPROCS(0))
		fmt.Printf("среднее время выполнения запроса: %f ms\n", float64(wp.GetAverageRequestTime())/float64(time.Millisecond))
		fmt.Println("ФИНИШ")

	}

	// проверяем, как завершился контекст
	// этот код просто для того, чтобы показать что я умею с этим работать
	if ctxCancel.Err() == context.Canceled {
		fmt.Println("Прервали общую работу")
	}
}

// таймер для отсчета времени.
func timer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	sec := 1
	for {
		select {
		case <-ticker.C:
			fmt.Printf("%d second\n", sec)
			sec++
		}
	}
}
