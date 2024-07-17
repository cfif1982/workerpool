package rcvrchan

import (
	"bufio"
	"fmt"
	"os"
)

type TaskReceiverChan struct {
	lineCH chan string
}

func NewTaskReceiverChan(filename string) *TaskReceiverChan {

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return nil
	}

	defer file.Close()

	// создаем объект
	// имеет ли здесь смысл создавать буферизированный канал? по идее можно создавть буфер по количеству воркеров
	// по идее это должно уменьшить количество блокировок
	t := &TaskReceiverChan{
		lineCH: make(chan string),
	}

	// Чтение файла и отправка строк в канал
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// отправляем строки в канал
			t.lineCH <- scanner.Text()
		}
		// по завершении считывания всех строк закрываем канал
		close(t.lineCH)

		// если при чтении произошла ошибка, выводим сообщение об этом
		if err := scanner.Err(); err != nil {
			fmt.Println("Ошибка при чтении файла:", err)
		}

	}()

	return t

}

// читаем следующую строку из канала
func (t *TaskReceiverChan) GetTask() (string, bool) {

	// читаем строку из канала
	str, ok := <-t.lineCH

	// Канал закрыт
	if !ok {
		return "", false
	}

	return str, true
}
