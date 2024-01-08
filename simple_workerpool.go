package main

import (
	"fmt"
	"time"
)

type Worker struct {
	WorkerCount int
	JobsCount   int
	Jobs        chan int
	Results     chan int
}

func NewWorkerPool(workerCount int, jobsCount int) *Worker {
	return &Worker{
		WorkerCount: workerCount,
		JobsCount:   jobsCount,
		Jobs:        make(chan int),
		Results:     make(chan int),
	}
}

func main() {

	pool := NewWorkerPool(10, 100)

	for w := 0; w < pool.WorkerCount; w++ {
		go func(id int) {
			for j := range pool.Jobs {

				pool.Results <- j * 2
				time.Sleep(time.Second)
			}
		}(w)
	}

	for j := 0; j < pool.JobsCount; j++ {
		go func(j int) {
			pool.Jobs <- j
		}(j)

	}

	for r := 0; r < pool.JobsCount; r++ {
		fmt.Println(<-pool.Results)
	}
}
