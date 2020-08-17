package gotaskengine

import "sync/atomic"

// FuncWork defines a function could deal parts.
type FuncWork func(part Part)

// Worker - worker take parts from conveyor and handle them.
type IWorker interface {
	Working()
	Stop()
}

// TWorker is the implementation of IWorker interface.
type TWorker struct {
	taskPool    IConveyor
	taskHandler FuncWork
	status      int32
}

func (w *TWorker) Working() {
	swapped := atomic.CompareAndSwapInt32(&w.status, StatusNew, StatusRun)
	if !swapped {
		return
	}

	for {
		part, ok := w.taskPool.GetPart()
		if !ok {
			atomic.StoreInt32(&w.status, StatusStop)
		}

		w.taskHandler(part)

		if atomic.LoadInt32(&w.status) == StatusStop {
			return
		}
	}
}

func (w *TWorker) Stop() {
	atomic.CompareAndSwapInt32(&w.status, StatusRun, StatusStop)
}

func NewWorker(conveyor IConveyor, handler FuncWork) *TWorker {
	return &TWorker{
		taskPool:    conveyor,
		taskHandler: handler,
		status:      StatusNew,
	}
}
