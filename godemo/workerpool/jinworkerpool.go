package workerpool

type JinWorkerPool interface {
	Submit(func())
}

