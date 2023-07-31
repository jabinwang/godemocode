package workerpool

type WorkerPool interface {
	Submit(func())
	SubmitWait(func())
	Stop()
	StopWait()
}

type workerPool struct {
	maxWorkers int
	taskQueue chan func()
	workerQueue chan func()
	waitingQueue Deque
	waitingCount int
}

func newPool(maxWorkers int) WorkerPool {
	p := &workerPool{
		maxWorkers: maxWorkers,
		taskQueue: make(chan func()),
		workerQueue: make(chan func()),
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
	<- doneChan
}

func (w *workerPool) Stop() {
	panic("implement me")
}

func (w *workerPool) StopWait() {
	panic("implement me")
}

func (w *workerPool) dispatch() {

}
