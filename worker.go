package gotaskengine

import "sync"

// Worker  worker take parts from conveyor and handle them
type Worker interface {
	Working(c <-chan interface{}, group *sync.WaitGroup)
}

type FuncWorker func(c <-chan interface{}, group *sync.WaitGroup)

func (f FuncWorker) Working(c <-chan interface{}, group *sync.WaitGroup) {
	go f(c, group)
}
