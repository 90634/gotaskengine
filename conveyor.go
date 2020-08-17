package gotaskengine

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ConveyorI receives parts, and workers take parts from it.
type IConveyor interface {

	// PutPart puts parts on conveyor belt, if a conveyor is full, return ErrLineIsFull.
	PutPart(p Part, duration time.Duration) error

	// GetPart return a part and the status of conveyor if closed.
	GetPart() (Part, bool)

	// Run lets the conveyor rolling.
	Run()

	// Stop stops a conveyor.
	Stop()
}

type TConveyor struct {
	status       int32          // status of the conveyor.
	rwmutex      sync.RWMutex   // keep multi-thread safety.
	handler      FuncWork       // handler defines how to handle a part on the conveyor
	workerMaxCnt int            // max limit of worker.
	workerMinCnt int            // min limit of worker.
	workersC     chan IWorker   // hold all running worker.
	StopSignC    chan struct{}  // receive stop signal.
	pipeline     chan Part      // a part channel.
	checkTime    *time.Ticker   // for auto-adjust worker's count.
	wg           sync.WaitGroup // for wait all workers are stopped.
}

// ErrLineIsFull conveyor is full, can not puts any part. Client judges whether to add more workers according to the CPU's load.
var ErrLineIsFull = errors.New("too many tasks, add a worker please")
var ErrLineStopped = errors.New("the conveyor is stopped")

func (c *TConveyor) PutPart(p Part, duration time.Duration) error {
	c.rwmutex.RLock()
	defer c.rwmutex.RUnlock()
	if c.status == StatusStop {
		return ErrLineStopped
	}

	tm := time.NewTimer(duration)
	defer tm.Stop()

	select {
	case c.pipeline <- p:
	case <-tm.C:
		return ErrLineIsFull
	}
	return nil
}

func (c *TConveyor) GetPart() (Part, bool) {
	p, ok := <-c.pipeline
	return p, ok
}

func (c *TConveyor) Run() {
	c.rwmutex.Lock()
	if c.status != StatusNew {
		c.rwmutex.Unlock()
		return
	}
	c.status = StatusRun
	c.rwmutex.Unlock()

	w := NewWorker(c, c.handler)
	go w.Working()
	c.workersC <- w
	c.wg.Add(1)
	for {
		select {
		case <-c.StopSignC:
			for w := range c.workersC {
				w.Stop()
				c.wg.Done()
				return
			}
		case <-c.checkTime.C:
			if len(c.pipeline) > 0 && len(c.workersC) < c.workerMaxCnt {
				w := NewWorker(c, c.handler)
				go w.Working()
				c.workersC <- w
				c.wg.Add(1)
				fmt.Println("add 1")
			} else if len(c.pipeline) == 0 && len(c.workersC) > c.workerMinCnt {
				w := <-c.workersC
				w.Stop()
				c.wg.Done()
				fmt.Println("sub 1")
			}
		}
	}
}

func (c *TConveyor) Stop() {
	c.rwmutex.Lock()
	if c.status != StatusRun {
		c.rwmutex.Unlock()
		return
	}
	c.status = StatusStop
	c.rwmutex.Unlock()

	c.checkTime.Stop()
	close(c.StopSignC)

	// wait all workers of the conveyor exit.
	c.wg.Wait()
	close(c.workersC)
	close(c.pipeline)
}

// NewConveyor create a conveyor instance.
func NewConveyor(pipeCap int, handler FuncWork, maxWorkerCnt, minWokerCnt int, interval time.Duration) IConveyor {
	c := new(TConveyor)
	c.status = StatusNew
	c.pipeline = make(chan Part, pipeCap)
	c.handler = handler
	c.workerMinCnt = maxWorkerCnt
	c.workerMinCnt = minWokerCnt
	c.workersC = make(chan IWorker, maxWorkerCnt)
	c.StopSignC = make(chan struct{})
	c.checkTime = time.NewTicker(interval)
	return c
}
