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
		// todo: I must test it_ is range chan, it can always wait
		if f == nil {
			w.pool.workerCache.Put(w)
			return
		}
		f()
		w.pool.PutWorker(w)
		w.pool.decRunning()
	}
}
