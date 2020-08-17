package gotaskengine

import (
	"errors"
	"sync"
	"sync/atomic"
)

type IFactory interface {
	// AddLine add conveyor to the factory.
	AddLine(name string, c IConveyor) error

	// Run lets all conveyors running.
	Run()

	// Stop lets all conveyors stop.
	Stop()

	// GetLine
	GetLine(name string) IConveyor
}

// emptyFactory a instance of Factory interface
type TFactory struct {
	status int32
	lines  map[string]IConveyor // it holds all conveyor
	wg     sync.WaitGroup       // for wait all workers are stopped.
}

var ErrUnallowed = errors.New("the factory is not allowed to add new conveyor")

func (f *TFactory) AddLine(name string, c IConveyor) error {
	if atomic.LoadInt32(&f.status) != StatusNew {
		return ErrUnallowed
	}

	f.lines[name] = c
	return nil
}

func (f *TFactory) GetLine(name string) IConveyor {
	return f.lines[name]
}

func (f *TFactory) Run() {
	swapped := atomic.CompareAndSwapInt32(&f.status, StatusNew, StatusRun)
	if !swapped {
		return
	}

	for _, c := range f.lines {
		go c.Run()
		f.wg.Add(1)
	}
}

func (f *TFactory) Stop() {
	swapped := atomic.CompareAndSwapInt32(&f.status, StatusRun, StatusStop)
	if !swapped {
		return
	}

	for _, l := range f.lines {
		go func(c IConveyor) {
			c.Stop()
			f.wg.Done()
		}(l)
	}

	f.wg.Wait()
}

func NewFactory() IFactory {
	f := new(TFactory)
	f.lines = make(map[string]IConveyor)
	return f
}
