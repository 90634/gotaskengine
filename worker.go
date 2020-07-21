package gotaskengine

// FuncDone define a function which is called by worker when a worker have done it's work
type FuncDone func()

// Worker  worker take parts from conveyor and handle them
type Worker interface {
	Working(part Part, done FuncDone, next Conveyor)
}

// FuncWorker is the implementation of Worker interface
type FuncWorker func(part Part, done FuncDone, next Conveyor)

func (f FuncWorker) Working(part Part, done FuncDone, next Conveyor) {
	go f(part, done, next)
}
