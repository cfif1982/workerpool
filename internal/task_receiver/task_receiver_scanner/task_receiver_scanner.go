package rcvrscan

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type TaskReceiverScanner struct {
	scanner *bufio.Scanner
	mu      sync.Mutex
}

func NewTaskReceiverFile(filename string) *TaskReceiverScanner {

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return nil
	}

	defer file.Close()

	// создаем сканер
	scanner := bufio.NewScanner(file)

	return &TaskReceiverScanner{
		scanner: scanner,
	}
}

// читаем следующую строку из файла
func (t *TaskReceiverScanner) GetTask() (string, bool) {

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.scanner.Scan() {
		return t.scanner.Text(), true
	}

	return "", false
}
