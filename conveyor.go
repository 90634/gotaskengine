package gotaskengine

import (
	"errors"
	"sync"
	"time"
)

// Conveyor put parts on it, and workers take parts from it
type Conveyor interface {

	// PutPart put parts on conveyor belt, if the conveyor is full, return ErrLineIsFull
	PutPart(p Part, duration time.Duration) error

	// Run lets the conveyor rolling
	Run()

	// Stop stop the conveyor and wait workers finish their work
	Stop()

	// Next return the only next Conveyor which parts from this flows to the next.
	Next() Conveyor
}

type emptyConveyor struct {
	running    bool
	mutex      sync.Mutex
	group      sync.WaitGroup
	worker     Worker
	next       Conveyor         // a worker will have handled a part, then put the part to the only next conveyor.
	pipeline   chan interface{} // a part channel.
	workersC   chan struct{}
	emptySignC chan struct{}
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
		<-c.workersC
	}

	go func() {
		for p := range c.pipeline {
			c.workersC <- struct{}{} // blocked here
			c.group.Add(1)
			c.worker.Working(p, done, c.next)
		}
		c.emptySignC <- struct{}{}
	}()
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

	if !c.running {
		return
	}

	c.running = false

	// stop receive parts
	close(c.pipeline)

	// protect c.group, can't call c.group.Wait at this time.
	<-c.emptySignC

	// wait all workers of the conveyor finish.
	c.group.Wait()
}

func (c *emptyConveyor) Next() Conveyor {
	return c.next
}

// NewConveyor create an emptyConveyor.
// cap is the capacity of the Conveyor.
// max is the maximum number of workers.
func NewConveyor(cap int, w Worker, maxWorkers int, next Conveyor) Conveyor {
	c := new(emptyConveyor)

	c.pipeline = make(chan interface{}, cap)
	c.worker = w
	c.workersC = make(chan struct{}, maxWorkers)
	c.next = next

	c.emptySignC = make(chan struct{}, 1)

	return c
}
