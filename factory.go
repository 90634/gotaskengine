package gotaskengine

import (
	"errors"
	"sync"
)

type Factory interface {
	// AddLine add conveyor to the factory.
	AddLine(c Conveyor) error

	// Run lets all conveyors running
	Run()

	// Stop lets all conveyors stop
	Stop()
}

// emptyFactory a instance of Factory interface
type emptyFactory struct {
	// it holds all conveyor
	lines   Graph
	running bool
	mutex   sync.Mutex
}

var ErrFactoryIsRunning = errors.New("the factory is running")

func (e *emptyFactory) AddLine(c Conveyor) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.running {
		return ErrFactoryIsRunning
	}

	var child *Node
	node := &Node{value: c, child: nil, parents: []*Node{}}

	if c.Next() != nil {
		child = &Node{value: c.Next(), child: nil, parents: []*Node{}}
		child.parents = append(child.parents, node)
		e.lines.addNode(child)
		node.child = child
	}
	e.lines.addNode(node)
	return nil
}

func (e *emptyFactory) Run() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.running {
		return
	}

	e.lines.makeIndexes()
	e.lines.runFromLeaves()

	e.running = true
}

func (e *emptyFactory) Stop() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.running {
		return
	}
	e.running = false

	e.lines.stopFromRoot()
}

func NewFactory() Factory {
	f := new(emptyFactory)
	return f
}

var ErrFactoryRunning = errors.New("factory is already running")

// defaultFactory
var defaultFactory = new(emptyFactory)

func AddLine(c Conveyor) error {
	return defaultFactory.AddLine(c)
}

func Run() {
	defaultFactory.Run()
}

func Stop() {
	defaultFactory.Stop()
}
