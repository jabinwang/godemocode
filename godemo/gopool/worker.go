package gopool

import (
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var workerPool sync.Pool

func init()  {
	workerPool.New = newWorker
}

type worker struct {
	pool *pool
}

func newWorker() interface{} {
	return &worker{}
}

func (w *worker) run()  {
	go func() {
		for  {
			var t *task
			w.pool.taskLock.Lock()
			if w.pool.taskHead != nil {
				t = w.pool.taskHead
				w.pool.taskHead = w.pool.taskHead.next
				atomic.AddInt32(&w.pool.taskCount, -1)
			}
			if t == nil {
				w.close()
				w.pool.taskLock.Unlock()
				w.Recycle()
				return
			}
			w.pool.taskLock.Unlock()
			func() {
				defer func() {
					if r:=recover(); r!= nil {
						if w.pool.panicHandler != nil {
							w.pool.panicHandler(t.ctx, r)
						} else {
							msg := fmt.Sprintf("jinpool: panic in pool %v: %s", r, debug.Stack())
							print(msg)
						}
					}
				}()
				t.f()
			}()
			t.Recycle()
		}
	}()
}

func (w *worker) close()  {
	w.pool.decWorkerCount()
}

func (w *worker) zero()  {
	w.pool = nil
}

func (w *worker) Recycle()  {
	w.zero()
	workerPool.Put(w)
}