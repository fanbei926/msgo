package mspool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var DefaultTime int64 = 3
var ErrInvaildCap = errors.New("cap can not less than 0")
var ErrInvaildExpire = errors.New("expire can not less than 0")
var ErrorHasClosed = errors.New("pool has bean released")

type sig struct{}

type Pool struct {
	// capacity
	cap int32
	// the number of running worker
	running int32
	// the number of idle workers
	workers []*Worker
	// expire time in seconds
	expire time.Duration
	// release signal
	release chan sig
	// lock
	lock sync.Mutex
	// sync.once for release
	once sync.Once
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultTime)
}

func NewTimePool(cap int, expire int64) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrInvaildCap
	}

	if expire <= 0 {
		return nil, ErrInvaildExpire
	}

	pool := &Pool{
		cap:     int32(cap),
		release: make(chan sig, 1),
		expire:  time.Duration(expire) * time.Second,
	}

	return pool, nil
}

func (pool *Pool) Submit(task func()) error {
	if len(pool.release) > 0 {
		return ErrorHasClosed
	}
	worker := pool.GetWorker()
	worker.task <- task
	pool.incRunning()
	return nil
}

func (pool *Pool) GetWorker() *Worker {
	idleWorkers := pool.workers
	n := len(idleWorkers) - 1
	// if pool has idle workers, return a worker
	if n >= 0 {
		pool.lock.Lock()
		worker := idleWorkers[n]
		idleWorkers[n] = nil
		pool.workers = idleWorkers[:n]
		pool.lock.Unlock()
		return worker
	}
	// if the number of running workers are less than cap, then create a new worker and return
	if pool.running < pool.cap {
		worker := &Worker{
			pool: pool,
			task: make(chan func(), 1),
		}
		worker.run()
		return worker
	}
	// if not, loop and wait for a new worker
	for {
		pool.lock.Lock()
		idleWorkers := pool.workers
		n := len(idleWorkers) - 1
		if n < 0 {
			pool.lock.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}

		worker := idleWorkers[n]
		idleWorkers[n] = nil
		pool.workers = idleWorkers[:n]
		pool.lock.Unlock()
		return worker
	}
}

func (pool *Pool) incRunning() {
	atomic.AddInt32(&pool.running, 1)
}

func (pool *Pool) decRunning() {
	atomic.AddInt32(&pool.running, -1)
}

func (pool *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	pool.lock.Lock()
	pool.workers = append(pool.workers, w)
	pool.lock.Unlock()
}

func (pool *Pool) Release() {
	pool.once.Do(func() {
		pool.lock.Lock()
		for i, w := range pool.workers {
			w.task = nil
			w.pool = nil
			pool.workers[i] = nil
		}
		pool.workers = nil
		pool.lock.Unlock()
		pool.release <- sig{}
	})
}

func (pool *Pool) IsClosed() bool {
	return len(pool.release) <= 0
}

func (pool *Pool) Restart() bool {
	if pool.IsClosed() {
		return true
	}
	_ = <-pool.release
	return true
}
