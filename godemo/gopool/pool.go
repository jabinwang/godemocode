package gopool

import (
	"context"
	"sync"
	"sync/atomic"
)

//默认值
const name = "default"
const capForRoutine = 32

type Pool interface {
	Go(func())
	SetPanicHandler(f func(context.Context, interface{}))
}

var taskPool sync.Pool

func init() {
	taskPool.New = newTask
}

type task struct {
	ctx  context.Context
	f    func()
	next *task
}

func (t *task) zero() {
	t.ctx = nil
	t.f = nil
	t.next = nil
}

func (t *task) Recycle() {
	t.zero()
	taskPool.Put(t)
}

func newTask() interface{} {
	return &task{}
}

type pool struct {
	taskHead     *task
	taskTail     *task
	taskCount    int32
	workerCount  int32
	taskLock     sync.Mutex
	panicHandler func(context.Context, interface{})
}

func (p *pool) Go(f func()) {
	p.ctxGo(context.Background(), f)
}

func (p *pool) ctxGo(ctx context.Context, f func()) {
	t := taskPool.Get().(*task)
	t.ctx = ctx
	t.f = f
	p.taskLock.Lock()
	if p.taskHead == nil {
		p.taskHead = t
		p.taskTail = t
	} else {
		p.taskTail.next = t
		p.taskTail = t
	}
	p.taskLock.Unlock()
	atomic.AddInt32(&p.taskCount, 1)
	if atomic.LoadInt32(&p.taskCount) >= 1 && p.WorkerCount() < 100 ||
				p.WorkerCount() == 0 {
		p.incWorkerCount()
		w := workerPool.Get().(*worker)
		w.pool = p
		w.run()
	}

}

func (p *pool) SetPanicHandler(f func(context.Context, interface{})) {
	p.panicHandler = f
}

func (p *pool) WorkerCount() int32 {
	return atomic.LoadInt32(&p.workerCount)
}

func (p *pool) incWorkerCount() {
	atomic.AddInt32(&p.workerCount, 1)
}

func (p *pool) decWorkerCount() {
	atomic.AddInt32(&p.workerCount, -1)
}

func newPool() Pool {
	p := &pool{
	}
	return p
}
