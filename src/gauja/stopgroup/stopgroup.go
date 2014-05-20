package stopgroup

import (
	"sync"
)

type StopGroup interface {
	Stopped() <-chan struct{}
	Stop()
}

type stopGroup struct {
	stopOnce   *sync.Once
	stopped    chan struct{}
	stopAction func()
}

func New(stopAction func()) StopGroup {
	return stopGroup{
		stopOnce:   new(sync.Once),
		stopped:    make(chan struct{}),
		stopAction: stopAction,
	}
}

func (sg stopGroup) Stopped() <-chan struct{} {
	return sg.stopped
}

func (sg stopGroup) Stop() {
	sg.stopOnce.Do(func() {
		close(sg.stopped)
		go sg.stopAction()
	})
}

func IsStopped(sg StopGroup) bool {
	select {
	case <-sg.Stopped():
		return true
	default:
		return false
	}
}
