package mspool

import "time"

type Worker struct {
	pool *Pool
	// tasks
	task chan func()
	// last time ?
	lastTime time.Time
}

func (w *Worker) run() {
	go w.running()
}

func (w *Worker) running() {
	for f := range w.task {
		if f == nil {
			return
		}
		f()
		w.pool.PutWorker(w)
		w.pool.decRunning()
	}
}
