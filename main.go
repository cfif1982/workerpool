package main

import (
	"fmt"
	"time"
	ressaver "training/internal/result_saver"                   // путь к вашему пакету examples
	trchan "training/internal/task_receiver/task_receiver_chan" // путь к вашему пакету examples
	workerpool "training/internal/worker_pool"                  // путь к вашему пакету examples
)

const countTasks = 10
const countWorkers = 3
const totalWorkTimeOut = 10 // таймаут общего времени работы всех воркеров

func main() {
	// Вызов примера MainWorkerpool
	fmt.Println("=== MainWorkerpool() ===")

	// создаем task receiver
	tr := trchan.NewTaskReceiverChan("addr.txt")

	// создаем result saver
	rs := ressaver.NewResultReceiver()
	defer rs.Close() // закрываем канал в result saver

	wp := workerpool.NewWorkerPool(countTasks, countWorkers, totalWorkTimeOut, tr, rs)

	// запускаем таймер для отображения счетчика времени
	go timer()

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
