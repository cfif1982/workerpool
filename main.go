package main

import (
	"fmt"
	"time"

	ressaver "github.com/cfif1982/workerpool/internal/result_saver"
	trchan "github.com/cfif1982/workerpool/internal/task_receiver/task_receiver_chan"
	workerpool "github.com/cfif1982/workerpool/internal/worker_pool"
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
