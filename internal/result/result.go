package result

import "time"

type Result struct {
	Body []byte
	Time time.Duration
}

func NewResult(body []byte, t time.Duration) *Result {
	return &Result{
		Body: body,
		Time: t,
	}
}
