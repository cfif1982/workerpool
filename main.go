package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	rschan "github.com/cfif1982/workerpool/internal/result_saver/result_saver_chan"
	trfile "github.com/cfif1982/workerpool/internal/task_receiver/task_receiver_file"
	workerpool "github.com/cfif1982/workerpool/internal/worker_pool"
)

const countWorkers = 10 // начальное количество воркеров

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
	// возвращается сигнальный канал о завершении работы  worker pool
	wpDoneCh := wp.Start(ctxCancel)

	select {
	// следим за сигнальным каналом
	case <-sigChan:
		fmt.Println("Grasefull Shutdown start")
		cancel() // отменяем работу всех воркеров при помощи контекста

	// если работа worker pool завершена, то завершаем программу
	case <-wpDoneCh:
		return
	}

	// проверяем, как завершился контекст
	switch ctxCancel.Err() {
	case context.Canceled:
		fmt.Println("Прервали общую работу")
	}

}

// таймер для отсчета времени
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
