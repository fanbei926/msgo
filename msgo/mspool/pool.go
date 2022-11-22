package mspool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var DefaultTime int64 = 5
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
	once         sync.Once
	workerCache  sync.Pool
	con          *sync.Cond
	PanicHandler func()
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultTime)
}

func NewPoolConf() (*Pool, error) {
	cap, ok := config.Conf.Pool["cap"]
	if !ok {
		return nil, errors.New("cap config not set")
	}
	return NewTimePool(int(cap.(int64)), DefaultTime)
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

	pool.con = sync.NewCond(&pool.lock)
	go pool.expireWorkers()
	return pool, nil
}

func (pool *Pool) Submit(task func()) error {
	if len(pool.release) > 0 {
		return ErrorHasClosed
	}
	worker := pool.GetWorker()
	worker.task <- task

	return nil
}

func (pool *Pool) GetWorker() *Worker {
	// below is the 3 states
	pool.lock.Lock()
	idleWorkers := pool.workers
	n := len(idleWorkers) - 1
	// 1. if pool has idle workers, return a worker
	if n >= 0 {
		worker := idleWorkers[n]
		idleWorkers[n] = nil
		pool.workers = idleWorkers[:n]
		pool.lock.Unlock()
		return worker
	}
	// 2. if the number of running workers are less than cap, then create a new worker and return
	// todo: shall we append the new worker into the slice? -- implement in PutWorker()
	if pool.running < pool.cap {
		pool.lock.Unlock()
		pool.workerCache.New = func() any {
			return &Worker{
				pool: pool,
				task: make(chan func(), 1),
			}
		}

		cache := pool.workerCache.Get()
		var worker *Worker
		if cache == nil {
			worker = &Worker{
				pool: pool,
				task: make(chan func(), 1),
			}
		} else {
			worker = cache.(*Worker)
		}

		worker.run()
		return worker
	}
	pool.lock.Unlock()

	// 3. if not, loop and wait for a new worker
	return pool.waitIdleWorker()
}

func (pool *Pool) waitIdleWorker() *Worker {
	pool.lock.Lock()
	pool.con.Wait()
	fmt.Println("得到通知，有空闲worker")
	idleWorkers := pool.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		pool.lock.Unlock()
		if pool.running < pool.cap {
			pool.workerCache.New = func() any {
				return &Worker{
					pool: pool,
					task: make(chan func(), 1),
				}
			}

			cache := pool.workerCache.Get()
			var worker *Worker
			if cache == nil {
				worker = &Worker{
					pool: pool,
					task: make(chan func(), 1),
				}
			} else {
				worker = cache.(*Worker)
			}

			worker.run()
			return worker
		}
		pool.waitIdleWorker()
	}

	// get the latest worker
	worker := idleWorkers[n]
	idleWorkers[n] = nil
	pool.workers = idleWorkers[:n]
	pool.lock.Unlock()
	return worker
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
	pool.con.Signal() // notify
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
	return len(pool.release) > 0
}

func (pool *Pool) Restart() bool {
	if pool.IsClosed() {
		return true
	}
	_ = <-pool.release
	go pool.expireWorkers()
	return true
}

func (pool *Pool) expireWorkers() {
	tick := time.NewTicker(pool.expire)

	for range tick.C {
		if pool.IsClosed() {
			break
		}
		pool.lock.Lock()
		n := len(pool.workers) - 1
		idleWorkers := pool.workers
		if n >= 0 {
			var clearN = -1
			for i, w := range idleWorkers {
				if time.Now().Sub(w.lastTime) < pool.expire {
					break
				}
				// below it means that the w is an expired worker
				w.task <- nil
				clearN = i // all elements before I are expired, so we need to delete them
			}
			if clearN != -1 {
				fmt.Println(clearN)
				var tmp []*Worker
				if clearN >= len(pool.workers)-1 {
					tmp = idleWorkers
					pool.workers = idleWorkers[:0]
					fmt.Printf("%s , 1 clean workers %v\n", time.Now().String(), tmp)
				} else {
					tmp = idleWorkers[:clearN+1]
					pool.workers = idleWorkers[clearN+1:]
					fmt.Printf("%s , 2 clean workers %v\n", time.Now().String(), tmp)
				}

			}
		}
		pool.lock.Unlock()
	}
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(p.cap - p.running)
}
