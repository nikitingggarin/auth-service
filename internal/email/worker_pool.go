package email

type WorkerPool struct {
	workers chan struct{}
}

func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		workers: make(chan struct{}, maxWorkers),
	}
}

func (wp *WorkerPool) Acquire() {
	wp.workers <- struct{}{}
}

func (wp *WorkerPool) Release() {
	<-wp.workers
}
