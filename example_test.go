package gotaskengine

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type Toy struct {
	Shape  int
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
	toyDogLine := NewConveyor(64, 4)
	toyCatLine := NewConveyor(64, 4)

	toyDogLine.SetWorker(FuncWorker(toyDogWorker))
	toyCatLine.SetWorker(FuncWorker(toyCatWorker))

	toyFactory := NewFactory()
	toyFactory.AddLine(toyDogLine)
	toyFactory.AddLine(toyCatLine)

	toyFactory.Run()

	stopC := make(chan bool)
	go func() {
		i := 0
		for {
			// you can insert this task into a database and set the status "not complete",then
			// you can retrieve these "not complete" tasks from database, and put them back in the queue
			err := toyCatLine.PutPart(ToyCat{Toy{i, "Meow~"}}, time.Second*2)
			if errors.Is(err, ErrLineIsFull) {
				t.Error(err.Error())
				// here, you can use github.com/shirou/gopsutil to get CPU's load, if it's ok, you can add a worker and retry.
			}
			if errors.Is(err, ErrLineIsStop) {
				t.Log(" toyCatLine stoped")
				break
			}

			//time.Sleep(time.Second * 1)
			i++
		}
		fmt.Printf("toyCat----%d----\n", i-1)
		stopC <- true
	}()

	go func() {
		i := 0
		for {
			err := toyDogLine.PutPart(ToyDog{Toy{i, "Wang~"}}, time.Second*2)
			if errors.Is(err, ErrLineIsFull) {
				t.Error(err.Error())
				// here, you can use github.com/shirou/gopsutil to get CPU's load, if it's ok, you can add a worker and retry.
			}
			if errors.Is(err, ErrLineIsStop) {
				t.Log(" toyDogLine stoped")
				break
			}

			//time.Sleep(time.Second * 1)
			i++
		}
		fmt.Printf("toyDog----%d----\n", i-1)
		stopC <- true
	}()

	// should wait stop signal
	time.Sleep(time.Second * 5)
	t.Log("Factory stopped")
	toyFactory.Stop()
	<-stopC
	<-stopC
}

func toyDogWorker(part Part, done FuncDone) {
	defer done()

	dogPart := part.(ToyDog)
	time.Sleep(time.Second * 1)
	fmt.Println(dogPart.Shape, dogPart.Action)
	// do something
}

func toyCatWorker(part Part, done FuncDone) {
	defer done()

	catPart := part.(ToyCat)
	time.Sleep(time.Second * 1)
	fmt.Println(catPart.Shape, catPart.Action)
	// do something
}
