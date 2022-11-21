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
	w.pool.incRunning()
	go w.running()
}

func (w *Worker) running() {
	defer func() {
		fmt.Println("Running")
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
			return
		}
		f()
		w.pool.PutWorker(w)
	}
}
