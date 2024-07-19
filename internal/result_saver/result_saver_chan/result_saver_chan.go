package saverchan

import (
	"github.com/cfif1982/workerpool/internal/result"
)

type ResultSaverChan struct {
	resultCH chan *result.Result
}

func NewResultSaver() *ResultSaverChan {

	return &ResultSaverChan{
		resultCH: make(chan *result.Result),
	}
}

func (r *ResultSaverChan) SaveResult(result *result.Result) {
	r.resultCH <- result
}

func (r *ResultSaverChan) Close() {
	close(r.resultCH)
}

// читаем следующий результат из канала
func (r *ResultSaverChan) GetResult() (*result.Result, bool) {

	// читаем результат из канала
	res, ok := <-r.resultCH

	// Канал закрыт
	if !ok {
		return nil, false
	}

	return res, true
}
