package proxy

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type WorkerPool struct {
	workerCount int
	jobQueue    chan *Job
	workers     []*Worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

type Job struct {
	Request  *http.Request
	Response http.ResponseWriter
	done     chan struct{}
	ctx      context.Context
}

func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan *Job, queueSize),
		workers:     make([]*Worker, workerCount),
		ctx:         ctx,
		cancel:      cancel,
	}
	return pool
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workerCount; i++ {
		worker := NewWorker(i, p.jobQueue, &p.wg, p.ctx)
		p.workers[i] = worker
		p.wg.Add(1)
		worker.Start()
	}
}

func (p *WorkerPool) Submit(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	
	done := make(chan struct{})
	job := &Job{
		Request:  r.WithContext(ctx),
		Response: w,
		done:     done,
		ctx:      ctx,
	}
	
	select {
	case p.jobQueue <- job:
		select {
		case <-done:
		case <-ctx.Done():
		}
	case <-p.ctx.Done():
	}
	
	cancel() 
}

func (p *WorkerPool) Stop() {
	p.cancel()
	select {
	case <-p.ctx.Done():
	default:
		close(p.jobQueue)
	} 
	
	p.wg.Wait()
}

type Worker struct {
	ID       int
	jobQueue chan *Job
	wg       *sync.WaitGroup
	proxy    *Proxy
	ctx      context.Context
}

func NewWorker(id int, jobQueue chan *Job, wg *sync.WaitGroup, ctx context.Context) *Worker {
	return &Worker{
		ID:       id,
		jobQueue: jobQueue,
		wg:       wg,
		ctx:      ctx,
	}
}

func (w *Worker) SetProxy(proxy *Proxy) {
	w.proxy = proxy
}

func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()
		for {
			select {
			case job, ok := <-w.jobQueue:
				if !ok {
					return
				}
				if w.proxy != nil {
					w.processJob(job)
				}
				close(job.done)
			case <-w.ctx.Done():
				return
			}
		}
	}()
}

func (w *Worker) processJob(job *Job) {
	select {
	case <-job.ctx.Done():
		http.Error(job.Response, "Request cancelled or timed out", http.StatusGatewayTimeout)
		return
	default:
	}

	if w.proxy.targetURL != nil {
		proxy := w.proxy.createReverseProxy(w.proxy.targetURL)
		proxy.ServeHTTP(job.Response, job.Request)
	} else {
		w.proxy.handleTransparentProxy(job.Response, job.Request)
	}
}