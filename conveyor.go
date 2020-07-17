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
	stopC      chan struct{}
}

// ErrLineIsFull conveyor is full, can not puts any part. Client judges whether to add more workers according to the CPU's load.
var ErrLineIsFull = errors.New("too many tasks, add a worker please")
var ErrLineIsStop = errors.New("the conveyor is stopped")

func (c *emptyConveyor) Run() {
	c.mutex.Lock()
	if c.running {
		c.mutex.Unlock()
		return
	}
	c.running = true
	c.mutex.Unlock()

	done := func() {
		c.group.Done()
		<-c.workerChan
	}

	for p := range c.pipeline {
		c.workerChan <- struct{}{} // block here
		c.group.Add(1)
		c.worker.Working(p, done)
	}
	c.stopC <- struct{}{}
}

func (c *emptyConveyor) SetWorker(w Worker) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.worker = w
}

func (c *emptyConveyor) PutPart(p Part, duration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.running {
		return ErrLineIsStop
	}

	tm := time.NewTimer(duration)
	select {
	case c.pipeline <- p:
	case <-tm.C:
		return ErrLineIsFull
	}

	tm.Stop()
	return nil
}

func (c *emptyConveyor) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.running = false
	close(c.pipeline)
	// protect c.group, can't call c.group.Wait at this time.
	<-c.stopC
	c.group.Wait()
}

// NewConveyor create an emptyConveyor.
// cap is the capacity of the Conveyor.
// max is the maximum number of workers.
func NewConveyor(cap int, max int) Conveyor {
	c := new(emptyConveyor)
	c.pipeline = make(chan interface{}, cap)
	c.workerChan = make(chan struct{}, max)
	c.stopC = make(chan struct{})
	return c
}
