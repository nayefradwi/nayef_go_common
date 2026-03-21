package errors

type ResultRunner struct{ Error error }

func (r *ResultRunner) Do(fn func() error) {
	if r.Error != nil {
		return
	}
	r.Error = fn()
}

type ResultRunnerWithParam[T any] struct{ Error error }

func (r *ResultRunnerWithParam[T]) Do(param T, fn func(param T) error) {
	if r.Error != nil {
		return
	}
	r.Error = fn(param)
}
