package proxy

import (
	"net/http"
	"sync"
)

type WorkerPool struct {
	workerCount int
	jobQueue    chan *Job
	workers     []*Worker
	wg          sync.WaitGroup
}

type Job struct {
	Request  *http.Request
	Response http.ResponseWriter
	done     chan struct{}
}

func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan *Job, queueSize),
		workers:     make([]*Worker, workerCount),
	}
	return pool
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workerCount; i++ {
		worker := NewWorker(i, p.jobQueue, &p.wg)
		p.workers[i] = worker
		p.wg.Add(1)
		worker.Start()
	}
}

func (p *WorkerPool) Submit(w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	job := &Job{
		Request:  r,
		Response: w,
		done:     done,
	}
	
	p.jobQueue <- job
	
	<-done
}

func (p *WorkerPool) Stop() {
	close(p.jobQueue)
	
	p.wg.Wait()
}

type Worker struct {
	ID       int
	jobQueue chan *Job
	wg       *sync.WaitGroup
	proxy    *Proxy
}

func NewWorker(id int, jobQueue chan *Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		ID:       id,
		jobQueue: jobQueue,
		wg:       wg,
	}
}

func (w *Worker) SetProxy(proxy *Proxy) {
	w.proxy = proxy
}

func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()
		for job := range w.jobQueue {
			// Process the job
			if w.proxy != nil {
				w.processJob(job)
			}
			
			close(job.done)
		}
	}()
}

func (w *Worker) processJob(job *Job) {
	if w.proxy.targetURL != nil {
		proxy := w.proxy.createReverseProxy(w.proxy.targetURL)
		proxy.ServeHTTP(job.Response, job.Request)
	} else {
		w.proxy.handleTransparentProxy(job.Response, job.Request)
	}
}