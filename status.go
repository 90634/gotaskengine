package gotaskengine

// status from StatusNew to StatusStopped, no loop.
const (
	StatusNew = iota
	StatusRun
	StatusStop
)
