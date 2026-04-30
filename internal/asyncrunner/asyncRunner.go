package asyncrunner

type asyncRunner struct{}

func New() *asyncRunner {
	return &asyncRunner{}
}

func (r *asyncRunner) Run(fn func()) {
	go fn()
}
