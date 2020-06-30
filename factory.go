package gotaskengine

import (
	"errors"
	"sync"
)

type Factory interface {
	// AddLine add conveyor to the factory.
	AddLine(c Conveyor)

	// Run lets all conveyors running
	Run()

	// Stop lets all conveyors stop
	Stop()
}

// emptyFactory a instance of Factory interface
type emptyFactory struct {
	lines   []Conveyor
	running bool
	mutex   sync.Mutex
}

func (e *emptyFactory) AddLine(c Conveyor) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.lines = append(e.lines, c)

	if e.running {
		c.Run()
	}
}

func (e *emptyFactory) Run() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.running {
		return
	}

	for _, line := range e.lines {
		line.Run()
	}

	e.running = true
}

func (e *emptyFactory) Stop() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, line := range e.lines {
		line.Stop()
	}

	e.running = false
}

func NewFactory() Factory {
	f := new(emptyFactory)
	return f
}

var ErrFactoryRunning = errors.New("factory is already running")

// defaultFactory
var defaultFactory = new(emptyFactory)

func AddLine(c Conveyor) {
	defaultFactory.AddLine(c)
}

func Run() {
	defaultFactory.Run()
}

func Stop() {
	defaultFactory.Stop()
}
