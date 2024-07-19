package main

import (
	"fmt"
	"time"

	rschan "github.com/cfif1982/workerpool/internal/result_saver/result_saver_chan"
	trchan "github.com/cfif1982/workerpool/internal/task_receiver/task_receiver_chan"
	workerpool "github.com/cfif1982/workerpool/internal/worker_pool"
)

const countWorkers = 3

func main() {

	// создаем result saver
	rs := rschan.NewResultSaver()
	defer rs.Close() // закрываем канал в result saver

	// создаем task receiver
	tr := trchan.NewTaskReceiverChan("addr.txt", rs)

	// создаем Worker Pool
	wp := workerpool.NewWorkerPool(countWorkers, tr, rs)

	// запускаем таймер для отображения счетчика времени
	go timer()

	// запускаем worker pool
	wp.Start()

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
