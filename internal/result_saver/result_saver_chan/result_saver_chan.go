package saverchan

import (
	"github.com/cfif1982/workerpool/internal/result"
)

// т.к. получатель результата не сделан, то нужно эмулировать его. Т.е. как будто он заберает результат из канала.
// для этого просто делаю буферизированный канал
const bufSize = 100 // размер буфера канала

// сохраняем результаты в канал.
type ResultSaverChan struct {
	resultCH chan *result.Result // канал, в который будем сохранять результат работы
}

// констурктор.
func NewResultSaver() *ResultSaverChan {
	return &ResultSaverChan{
		resultCH: make(chan *result.Result, bufSize), // создаем буферизированный канал для теста
	}
}

// сохраянем результат
// result - передаваемый результат
// ok - true если результат получен, false - если результат не получен (работа закончена).
// переделал - ок не нужен, т.к. в случае передачи указателя в канал, его zeroValue будет nil. По этому nil и можно отследить закрытие канала
func (r *ResultSaverChan) SaveResult(result *result.Result) {
	// если  работа закончена, то закрываем канал
	if result == nil {
		close(r.resultCH)

		return
	}

	// если работа еще не закончена, то отправляем данные в канал
	r.resultCH <- result
}
