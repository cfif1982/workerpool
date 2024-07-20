package trfile

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/cfif1982/workerpool/internal/task"
)

// получаем задачи из файла
type TaskReceiverFile struct {
	filename string          // название файла, из которого будем читать строки
	taskCH   chan *task.Task // канал, в который будем отсылать задачи
}

// конструктор
func NewTaskReceiverFile(filename string) *TaskReceiverFile {

	return &TaskReceiverFile{
		filename: filename,
		taskCH:   make(chan *task.Task), // TODO: нужно ли его делать буферизированным?
	}
}

// возвращает канал, по которому  будeт отсылаться задачи
func (t *TaskReceiverFile) GetTaskCH() chan *task.Task {

	return t.taskCH
}

// запускаем работу task receiver
// wg нужен для того, чтобы программа не закончилась раньше того, как закончится генерация задач
// ctx - для отменаы работы
func (t *TaskReceiverFile) Start(ctx context.Context, wg *sync.WaitGroup) {

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

		// сканер для работы с файлом
		scanner := bufio.NewScanner(file)

		// считываем строки из файла
		for scanner.Scan() {
			select {
			// следим за контекстом
			case <-ctx.Done():
				return

			// если контекст не отменен
			default:

				i++ // увеличиваем счетчик задач

				// создаем задачу и отправляем ее в канал
				t.taskCH <- task.NewTask(i, scanner.Text())
			}
		}

		// по завершении считывания всех строк закрываем канал
		close(t.taskCH)

		// если при чтении произошла ошибка, выводим сообщение об этом
		if err := scanner.Err(); err != nil {
			fmt.Println("Ошибка при чтении файла:", err)
		}

	}()
}
