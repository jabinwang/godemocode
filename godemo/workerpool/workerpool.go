package workerpool

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	idleTimeout = 2 * time.Second
)

type WorkerPool interface {
	Submit(func())
	SubmitWait(func())
	Stop()
	StopWait()
}

type workerPool struct {
	maxWorkers   int
	taskQueue    chan func()
	workerQueue  chan func()
	stoppedChan chan struct{}
	waitingQueue Deque
	waitingCount int32
	stopOnce sync.Once
	stopLock sync.Mutex
	stopped bool
	wait bool
}

func newPool(maxWorkers int) WorkerPool {
	p := &workerPool{
		maxWorkers:  maxWorkers,
		taskQueue:   make(chan func()),
		workerQueue: make(chan func()),
		stoppedChan: make(chan struct{}),
		waitingQueue: New(16),
	}
	go p.dispatch()
	return p
}

func (w *workerPool) Submit(task func()) {
	if task != nil {
		w.taskQueue <- task
	}
}

func (w *workerPool) SubmitWait(task func()) {
	if task == nil {
		return
	}
	doneChan := make(chan struct{})
	w.taskQueue <- func() {
		task()
		close(doneChan)
	}
	<-doneChan
}

func (w *workerPool) Stop() {
	w.stop(false)
}

func (w *workerPool) StopWait() {
	w.stop(true)
}

func (w *workerPool) stop(wait bool)  {
	w.stopOnce.Do(func() {
		w.stopLock.Lock()
		w.stopped = true
		w.stopLock.Unlock()
		w.wait = wait
		close(w.taskQueue)
	})
	<- w.stoppedChan
}

func (w *workerPool) dispatch() {
	defer close(w.stoppedChan)
	timeout := time.NewTimer(idleTimeout)
	var workerCount int
	var wg sync.WaitGroup
	var idle bool
Loop:
	for {
		if w.waitingQueue.Len() != 0 {
			if !w.processWaitingQueue() {
				break Loop
			}
			continue
		}
		select {
		case task, ok := <-w.taskQueue:
			if !ok {
				break Loop
			}
			select {
			case w.workerQueue <- task:
			default:
				if workerCount < w.maxWorkers {
					wg.Add(1)
					go workerRunner(task, w.workerQueue, &wg)
					workerCount++
				} else {
					w.waitingQueue.PushBack(task)
					atomic.StoreInt32(&w.waitingCount, int32(w.waitingQueue.Len()))
				}
				idle = false
			case <-timeout.C:
				if idle && workerCount > 0 {
					if w.killIdleWorker() {
						workerCount--
					}
				}
				idle = true
				timeout.Reset(idleTimeout)
			}
		}
	}
	
	if w.wait {
		w.runQueuedTasks()
	}

	for workerCount > 0 {
		w.workerQueue <- nil
		workerCount--
	}
	wg.Wait()
	timeout.Stop()
}

func workerRunner(task func(), workerQueue chan func(), wg *sync.WaitGroup) {
	for task != nil {
		task()
		task = <-workerQueue
	}
	wg.Done()
}

func (w *workerPool) processWaitingQueue() bool {
	select {
	case task, ok := <-w.taskQueue:
		if !ok {
			return false
		}
		w.waitingQueue.PushBack(task)
	case w.workerQueue <- w.waitingQueue.Front().(func()):
		w.waitingQueue.PopFront()
	}
	atomic.StoreInt32(&w.waitingCount, int32(w.waitingQueue.Len()))
	return true
}

func (w *workerPool) killIdleWorker() bool {
	select {
	case w.workerQueue <- nil:
		return true
	default:
		return false
	}
}

func (w *workerPool) runQueuedTasks() {
	for w.waitingQueue.Len() != 0 {
		w.workerQueue <- w.waitingQueue.PopFront().(func())
		atomic.StoreInt32(&w.waitingCount, int32(w.waitingQueue.Len()))
	}
}
