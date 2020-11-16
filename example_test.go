package gotaskengine

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type Toy struct {
	Id     int
	Action string
}

type ToyDog struct {
	Toy
	//other
}

var ToyFactory IFactory

func TestFactory(t *testing.T) {
	dogLegsLine := NewConveyor(16, FuncWork(dogLegsWorker), 8, 1, time.Second*5)
	dogBodyLine := NewConveyor(16, FuncWork(dogBodyWorker), 8, 1, time.Second*2)

	ToyFactory = NewFactory()

	err := ToyFactory.AddLine("dogBodyLine", dogBodyLine)
	if err != nil {
		t.Error(err)
	}

	err = ToyFactory.AddLine("dogLegsLine", dogLegsLine)
	if err != nil {
		t.Error(err)
	}

	ToyFactory.Run()

	//stopSignC := make(chan bool)

	go func() {
		i := 0
		for {
			err := ToyFactory.GetLine("dogBodyLine").PutPart(ToyDog{Toy{i, "Wang~"}}, time.Second*1)
			if errors.Is(err, ErrLineIsFull) {
				i--
				// here, you can use github.com/shirou/gopsutil to get CPU's load, if it's ok, you can add a worker and retry.
			}
			if errors.Is(err, ErrLineStopped) {
				t.Log(" dogBodyLine stoped")
				break
			}

			//time.Sleep(time.Second * 1)
			i++
			fmt.Printf("toyDog total :%d\n", i-1)
		}
	}()

	// wait stop signal
	time.Sleep(time.Second * 30)
	t.Log("Factory stopped")
	ToyFactory.Stop()
}

func dogBodyWorker(part Part) {
	dogPart := part.(ToyDog)
	time.Sleep(time.Second * 3)
	fmt.Printf("ToyDog %d body is ok\n", dogPart.Id)
	// do stuff here

	err := ToyFactory.GetLine("dogLegsLine").PutPart(dogPart, time.Second*5)
	if err != nil {
		fmt.Println("1111: ", err)
	}
}

func dogLegsWorker(part Part) {
	dogPart := part.(ToyDog)
	time.Sleep(time.Second * 2)
	fmt.Printf("ToyDog %d legs is ok, %s\n", dogPart.Id, dogPart.Action)
	// do stuff here
}
