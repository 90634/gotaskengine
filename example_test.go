package gotaskengine

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type Toy struct {
	Number int
	Action string
}

type ToyDog struct {
	Toy
	//other
}

func TestFactory(t *testing.T) {
	DogLegsLine := NewConveyor(64, FuncWorker(dogLegsWorker), 4, nil)
	DogBodyLine := NewConveyor(64, FuncWorker(dogBodyWorker), 8, DogLegsLine)

	toyFactory := NewFactory()
	_ = toyFactory.AddLine(DogBodyLine)
	_ = toyFactory.AddLine(DogLegsLine)

	toyFactory.Run()

	//stopSignC := make(chan bool)

	go func() {
		i := 0
		for {
			err := DogBodyLine.PutPart(ToyDog{Toy{i, "Wang~"}}, time.Second*2)
			if errors.Is(err, ErrLineIsFull) {
				t.Error(err)
				// here, you can use github.com/shirou/gopsutil to get CPU's load, if it's ok, you can add a worker and retry.
			}
			if errors.Is(err, ErrLineIsStop) {
				t.Log(" toyDogLine stoped")
				break
			}

			//time.Sleep(time.Second * 1)
			i++
		}
		fmt.Printf("toyDog total :%d\n", i-1)
	}()

	// should wait stop signal
	time.Sleep(time.Second * 10)
	t.Log("Factory stopped")
	toyFactory.Stop()
}

func dogBodyWorker(part Part, done FuncDone, next Conveyor) {
	defer done()

	dogPart := part.(ToyDog)
	time.Sleep(time.Second * 1)
	fmt.Printf("ToyDog %d body is ok\n", dogPart.Number)
	// do something

	err := next.PutPart(dogPart, time.Second*5)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func dogLegsWorker(part Part, done FuncDone, next Conveyor) {
	defer done()

	dogPart := part.(ToyDog)
	time.Sleep(time.Second * 1)
	fmt.Printf("ToyDog %d legs is ok, %s\n", dogPart.Number, dogPart.Action)
	// do something
}
