package mspool

import (
	"fmt"
	"time"
)

type Worker struct {
	pool *Pool
	// tasks
	task chan func()
	// last time ?
	lastTime time.Time
}

func (w *Worker) run() {
	go w.running()
	w.pool.incRunning()
}

func (w *Worker) running() {
	defer func() {
		w.pool.decRunning()
		w.pool.workerCache.Put(w)
		if err := recover(); err != nil {
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler()
			} else {
				fmt.Println("no handler")
			}
		}
		w.pool.con.Signal()
	}()
	for f := range w.task {
		// todo: I must test it_ is range chan, it can always wait
		if f == nil {
			w.pool.workerCache.Put(w)
			return
		}
		f()
		w.pool.PutWorker(w)
	}
}
