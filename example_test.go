package gotaskengine

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

type Toy struct {
	Shape  string
	Action string
}

type ToyDog struct {
	Toy
	//other
}

type ToyCat struct {
	Toy
	//other
}

func TestFactory(t *testing.T) {
	toyDogLine := NewConveyor(64)
	toyCatLine := NewConveyor(64)

	toyDogLine.AddWorker(FuncWorker(toyDogWorker), 2)
	toyCatLine.AddWorker(FuncWorker(toyCatWorker), 2)

	toyFactory := NewFactory()
	toyFactory.AddLine(toyDogLine)
	toyFactory.AddLine(toyCatLine)

	toyFactory.Run()

	go func() {
		for {
			// you can insert this task into a database and set the status "not complete",then
			// you can retrieve these "not complete" tasks from database, and put them back in the queue
			err := toyCatLine.PutPart(ToyCat{Toy{"four legs", "Meow～"}}, time.Second*2)
			if errors.Is(err, ErrLineIsFull) {
				t.Error(err.Error())
				// here, you can use github.com/shirou/gopsutil to get CPU's load, if it's ok, you can add a worker and retry.
			}

			time.Sleep(time.Second * 1)
			err = toyDogLine.PutPart(ToyDog{Toy{"two ears", "Wu~"}}, time.Second*2)
			if errors.Is(err, ErrLineIsFull) {
				t.Error(err.Error())
				// here, you can use github.com/shirou/gopsutil to get CPU's load, if it's ok, you can add a worker and retry.
			}
		}
	}()

	// should wait stop signal
	time.Sleep(time.Second * 30)
	toyFactory.Stop()
}

func toyDogWorker(c <-chan interface{}, group *sync.WaitGroup) {
	defer group.Done()

	for p := range c {
		dogPart := p.(ToyDog)
		fmt.Println(dogPart.Shape, dogPart.Action)
		// do something
		// In some cases，break the loop
	}
}

func toyCatWorker(c <-chan interface{}, group *sync.WaitGroup) {
	defer group.Done()

	for p := range c {
		catPart := p.(ToyCat)
		fmt.Println(catPart.Shape, catPart.Action)
		// do something
		// In some cases，break the loop
	}
}
