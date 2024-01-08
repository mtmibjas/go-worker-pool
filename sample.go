package main

import (
	"fmt"
	"sync"
)

// Worker represents a worker that executes tasks.
type Worker struct {
	ID     int
	TaskCh chan Task
	QuitCh chan struct{}
}

// Task represents a task that can be executed by a worker.
type Task func()

// NewWorker creates a new worker with the given ID.
func NewWorker(id int, wg *sync.WaitGroup) *Worker {
	worker := &Worker{
		ID:     id,
		TaskCh: make(chan Task),
		QuitCh: make(chan struct{}),
	}

	go worker.start(wg)
	return worker
}

// start starts the worker, listening for tasks and quitting when necessary.
func (w *Worker) start(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case task := <-w.TaskCh:
			task() // Execute the task.
		case <-w.QuitCh:
			return // Quit the worker.
		}
	}
}

// Stop stops the worker.
func (w *Worker) Stop() {
	close(w.QuitCh)
}

// WorkerPool represents a pool of workers.
type WorkerPool struct {
	Workers []*Worker
	TaskCh  chan Task
	WG      sync.WaitGroup
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
func NewWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		Workers: make([]*Worker, numWorkers),
		TaskCh:  make(chan Task),
	}

	for i := 0; i < numWorkers; i++ {
		pool.Workers[i] = NewWorker(i, &pool.WG)
		pool.WG.Add(1)
	}

	go pool.start()
	return pool
}

// start starts the worker pool, dispatching tasks to available workers.
func (p *WorkerPool) start() {
	defer p.WG.Wait()

	for {
		select {
		case task := <-p.TaskCh:
			// Dispatch task to an available worker.
			worker := p.getAvailableWorker()
			worker.TaskCh <- task
		}
	}
}

// getAvailableWorker finds an available worker or waits until one is available.
func (p *WorkerPool) getAvailableWorker() *Worker {
	for _, worker := range p.Workers {
		select {
		case <-worker.QuitCh:
			// Restart the worker if it has been stopped.
			worker = NewWorker(worker.ID, &p.WG)
			p.WG.Add(1)
			return worker
		default:
			if len(worker.TaskCh) == 0 {
				return worker
			}
		}
	}

	// If no available workers, wait for a worker to be available.
	for _, worker := range p.Workers {
		<-worker.TaskCh // Wait for a worker to be available.
		return worker
	}
}

// Stop stops the worker pool and all its workers.
func (p *WorkerPool) Stop() {
	for _, worker := range p.Workers {
		worker.Stop()
	}
	close(p.TaskCh)
}

func main() {
	// Create a worker pool with 3 workers.
	pool := NewWorkerPool(3)

	// Submit tasks to the worker pool.
	for i := 0; i < 10; i++ {
		index := i // Capture the loop variable.
		task := func() {
			fmt.Printf("Task %d executed by worker %d\n", index, pool.getAvailableWorker().ID)
		}
		pool.TaskCh <- task
	}

	// Stop the worker pool when all tasks are submitted.
	pool.Stop()
}
