package resultsaver

type ResultSaver struct {
	resultCH chan string
}

func NewResultReceiver() *ResultSaver {

	return &ResultSaver{
		resultCH: make(chan string),
	}
}

func (r *ResultSaver) SaveResult(str string) {
	r.resultCH <- str
}

func (r *ResultSaver) Close() {
	close(r.resultCH)
}
