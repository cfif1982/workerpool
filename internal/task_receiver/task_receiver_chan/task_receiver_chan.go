package rcvrchan

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/cfif1982/workerpool/internal/result"
	"github.com/cfif1982/workerpool/internal/task"
)

type ResultSaverI interface {
	SaveResult(result *result.Result)
}

type TaskReceiverChan struct {
	filename string
	taskCH   chan *task.Task
	rs       ResultSaverI
}

func NewTaskReceiverChan(filename string, rs ResultSaverI) *TaskReceiverChan {

	return &TaskReceiverChan{
		filename: filename,
		taskCH:   make(chan *task.Task),
		rs:       rs,
	}
}

// читаем следующую строку из канала
// wg нужен для того, чтобы программа не закончилась раньше того, как закончится генерация задач
func (t *TaskReceiverChan) Start(wg *sync.WaitGroup) {

	// Чтение файла и отправка строк в канал
	go func() {
		defer wg.Done() // по завершении чтения всего файла уменьшаем wg

		file, err := os.Open(t.filename)
		if err != nil {
			fmt.Println("Ошибка при открытии файла:", err)
			return
		}

		defer file.Close()

		i := 0 // счетчки задач

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {

			i++

			// создаем задачу
			task := task.NewTask(i, scanner.Text(), t.rs)

			// отправляем задачи в канал
			t.taskCH <- task
		}

		// по завершении считывания всех строк закрываем канал
		close(t.taskCH)

		// если при чтении произошла ошибка, выводим сообщение об этом
		if err := scanner.Err(); err != nil {
			fmt.Println("Ошибка при чтении файла:", err)
		}

	}()
}

// читаем следующую строку из канала
// возвращем true если канал открыт, false если канал закрыт
func (t *TaskReceiverChan) GetTask(ctx context.Context) (*task.Task, bool) {

	select {
	// следим за контекстом для отмены
	case <-ctx.Done():
		return nil, false

		// читаем задачу из канала
	case task, ok := <-t.taskCH:
		// Канал закрыт
		if !ok {
			return nil, false
		}

		return task, true
	}
}
