package gotaskengine

import (
	"errors"
	"sync"
	"time"
)

// Conveyor put parts on it, and workers take parts from it
type Conveyor interface {

	// SetWorker set the instance of Worker which work on the conveyor belt
	SetWorker(w Worker)

	// PutPart put parts on conveyor belt, if the conveyor is full, return ErrLineIsFull
	PutPart(p Part, duration time.Duration) error

	// Run lets the conveyor rolling
	Run()

	// Stop stop the conveyor and wait workers finish their work
	Stop()
}

type emptyConveyor struct {
	running    bool
	mutex      sync.Mutex
	group      sync.WaitGroup
	pipeline   chan interface{}
	worker     Worker
	workerChan chan struct{}
}

// ErrLineIsFull conveyor is full, can not puts any part. Client judges whether to add more workers according to the CPU's load.
var ErrLineIsFull = errors.New("too many tasks, add a worker please")
var ErrLineIsStop = errors.New("the conveyor is stopped")

func (e *emptyConveyor) Run() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.running {
		return
	}
	e.running = true

	go func() {
		for p := range e.pipeline {
			e.workerChan <- struct{}{} // block here
			e.group.Add(1)
			e.worker.Working(p, func() {
				<-e.workerChan
				e.group.Done()
			})
		}
	}()
}

func (e *emptyConveyor) SetWorker(w Worker) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.worker = w
}

func (e *emptyConveyor) PutPart(p Part, duration time.Duration) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if !e.running {
		return ErrLineIsStop
	}

	tm := time.NewTimer(duration)
	select {
	case e.pipeline <- p:
	case <-tm.C:
		return ErrLineIsFull
	}

	tm.Stop()
	return nil
}

func (e *emptyConveyor) Stop() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.running = false
	close(e.pipeline)
	e.group.Wait()
}

// NewConveyor create an emptyConveyor.
// cap is the capacity of the Conveyor.
// max is the maximum number of workers.
func NewConveyor(cap int, max int) Conveyor {
	c := new(emptyConveyor)
	c.pipeline = make(chan interface{}, cap)
	c.workerChan = make(chan struct{}, max)
	return c
}
