package saverchan

import (
	"github.com/cfif1982/workerpool/internal/result"
)

// сохраняем результаты в канал
type ResultSaverChan struct {
	resultCH chan *result.Result // канал, в который будем сохранять результат работы
}

// констурктор
func NewResultSaver() *ResultSaverChan {

	return &ResultSaverChan{
		resultCH: make(chan *result.Result, 100), // создаем буферизированный канал для теста
	}
}

// сохраянем результат
// result - передаваемый результат
// ok - true если результат получен, false - если результат не получен (работа закончена)
func (r *ResultSaverChan) SaveResult(result *result.Result, ok bool) {

	// если работа еще не закончена, то отправляем данные в канал
	if ok {
		r.resultCH <- result

		// если закончена работа, то закрываем канал
	} else {
		close(r.resultCH)
	}
}
