package utils

import (
	"github.com/schollz/progressbar/v3"
	"log"
	"sync"
)

type TaskOrchestrator struct {
	bar   *progressbar.ProgressBar
	wg    sync.WaitGroup
	mutex sync.Mutex
	sem   chan int
}

func NewTaskOrchestrator(bar *progressbar.ProgressBar, numberOfTasks int, maxConcurrentOperations int64) *TaskOrchestrator {
	task := TaskOrchestrator{
		bar: bar,
		sem: make(chan int, maxConcurrentOperations),
	}

	task.wg.Add(numberOfTasks)
	return &task
}

func (task *TaskOrchestrator) StartTask() {
	task.sem <- 1
}

func (task *TaskOrchestrator) Lock() {
	task.mutex.Lock()
}

func (task *TaskOrchestrator) Unlock() {
	task.mutex.Unlock()
}

func (task *TaskOrchestrator) FinishTask() {
	err := task.bar.Add(1)

	if err != nil {
		log.Printf("failed to update progress bar: %v", err)
	}

	<-task.sem
	task.wg.Done()
}

func (task *TaskOrchestrator) WaitForTasks() {
	task.wg.Wait()
}
