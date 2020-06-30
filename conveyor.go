package gotaskengine

import (
	"errors"
	"sync"
	"time"
)

// Conveyor put parts on it, and workers take parts from it
type Conveyor interface {

	// GetSelf return the instance
	GetSelf() chan interface{}

	// AddWorker add n workers on the conveyor belt
	AddWorker(w Worker, n int)

	// PutPart put parts on conveyor belt, if the conveyor is full, return ErrLineIsFull
	PutPart(p Part, duration time.Duration) error

	// Run lets the conveyor rolling
	Run()

	// Stop stop the conveyor and wait workers finish their work
	Stop()
}

type emptyConveyor struct {
	pipeline chan interface{}
	workers  []Worker
	running  bool
	group    sync.WaitGroup
	mutex    sync.Mutex
}

// ErrLineIsFull conveyor is full, can not puts any part. Client judges whether to add more workers according to the CPU's load.
var ErrLineIsFull = errors.New("too many tasks, add a worker please")

func (e *emptyConveyor) GetSelf() chan interface{} {
	return e.pipeline
}

func (e *emptyConveyor) Run() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for _, w := range e.workers {
		w.Working(e.pipeline, &e.group)
		e.group.Add(1)
	}

	e.running = true
}

func (e *emptyConveyor) AddWorker(w Worker, n int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for i := 0; i < n; i++ {
		e.workers = append(e.workers, w)
	}
	if e.running {
		w.Working(e.pipeline, &e.group)
	}
}

func (e *emptyConveyor) PutPart(p Part, duration time.Duration) error {
	tm := time.NewTimer(duration)
	select {
	// if e.pipeline is closed ,that ok?
	case e.pipeline <- p:
	case <-tm.C:
		return ErrLineIsFull
	}

	return nil
}

func (e *emptyConveyor) Stop() {
	close(e.pipeline)
	e.group.Wait()
}

func NewConveyor(cap int) Conveyor {
	c := new(emptyConveyor)
	c.pipeline = make(chan interface{}, cap)
	return c
}
